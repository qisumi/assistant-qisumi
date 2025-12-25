package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"assistant-qisumi/internal/logger"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.uber.org/zap"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ChatRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Tools            []Tool    `json:"tools,omitempty"`
	ToolChoice       string    `json:"tool_choice,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	ThinkingType     string    `json:"thinking_type,omitempty"`     // disabled, enabled, auto
	ReasoningEffort  string    `json:"reasoning_effort,omitempty"`  // low, medium, high, minimal
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolCallFunc `json:"function"`
}

type ToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

type Client interface {
	Chat(ctx context.Context, cfg Config, req ChatRequest) (*ChatResponse, error)
}

type HTTPClient struct {
	httpClient *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{httpClient: &http.Client{}}
}

func (c *HTTPClient) Chat(ctx context.Context, cfg Config, req ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	if ctx == nil {
		ctx = context.Background()
	}

	logger.Logger.Debug("LLM Client Chat请求开始",
		zap.String("model", req.Model),
		zap.String("base_url", cfg.BaseURL),
		zap.Int("messages_count", len(req.Messages)),
		zap.Int("tools_count", len(req.Tools)),
		zap.String("tool_choice", req.ToolChoice),
	)

	params, err := buildChatParams(req)
	if err != nil {
		logger.Logger.Error("构建Chat请求参数失败",
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	client := newOpenAIClient(cfg, c.httpClient)
	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		duration := time.Since(startTime)
		logger.Logger.Error("LLM API调用失败",
			zap.String("model", req.Model),
			zap.String("error", err.Error()),
			zap.Duration("duration", duration),
		)
		return nil, err
	}

	duration := time.Since(startTime)
	chatResp := fromOpenAIChatResponse(resp)

	logger.Logger.Info("LLM API调用成功",
		zap.String("model", req.Model),
		zap.Int("choices_count", len(chatResp.Choices)),
		zap.Duration("duration", duration),
	)

	// 记录响应详情（debug级别）- 包含完整内容
	if len(chatResp.Choices) > 0 {
		choice := chatResp.Choices[0]
		logger.Logger.Debug("LLM响应详情",
			zap.String("finish_reason", choice.FinishReason),
			zap.Int("content_length", len(choice.Message.Content)),
			zap.Int("tool_calls_count", len(choice.Message.ToolCalls)),
			zap.String("content", choice.Message.Content),
		)
		
		// 如果有工具调用，记录工具调用详情
		if len(choice.Message.ToolCalls) > 0 {
			for i, toolCall := range choice.Message.ToolCalls {
				logger.Logger.Debug("LLM工具调用详情",
					zap.Int("index", i),
					zap.String("tool_call_id", toolCall.ID),
					zap.String("tool_name", toolCall.Function.Name),
					zap.String("tool_arguments", toolCall.Function.Arguments),
				)
			}
		}
	}

	return chatResp, nil
}

func newOpenAIClient(cfg Config, httpClient *http.Client) openai.Client {
	opts := []option.RequestOption{}
	if cfg.APIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.APIKey))
	}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	if httpClient != nil {
		opts = append(opts, option.WithHTTPClient(httpClient))
	}
	return openai.NewClient(opts...)
}

func buildChatParams(req ChatRequest) (openai.ChatCompletionNewParams, error) {
	messages, err := toOpenAIMessages(req.Messages)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(req.Model),
		Messages: messages,
	}

	if req.Temperature != 0 {
		params.Temperature = openai.Float(req.Temperature)
	}
	if req.MaxTokens != 0 {
		params.MaxTokens = openai.Int(int64(req.MaxTokens))
	}
	if len(req.Tools) > 0 {
		tools, err := toOpenAITools(req.Tools)
		if err != nil {
			return openai.ChatCompletionNewParams{}, err
		}
		params.Tools = tools
	}
	if req.ToolChoice != "" {
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openai.String(req.ToolChoice),
		}
	}

	// 处理thinking_type和reasoning_effort参数
	// 注意：reasoning_effort是OpenAI较新的参数，当前SDK版本可能不完全支持
	// 这些参数已经存储在数据库中，前端可配置
	// TODO: 需要升级SDK或使用自定义HTTP客户端来传递这些参数
	// 目前暂时注释掉，等SDK支持后再启用
	_ = req.ThinkingType    // 暂时未使用，保留字段避免警告
	_ = req.ReasoningEffort // 暂时未使用，保留字段避免警告

	return params, nil
}

func toOpenAIMessages(messages []Message) ([]openai.ChatCompletionMessageParamUnion, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		converted, err := toOpenAIMessage(msg)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}

	return out, nil
}

func toOpenAIMessage(msg Message) (openai.ChatCompletionMessageParamUnion, error) {
	switch msg.Role {
	case "system":
		var system openai.ChatCompletionSystemMessageParam
		if msg.Content != "" {
			system.Content.OfString = openai.String(msg.Content)
		}
		if msg.Name != "" {
			system.Name = openai.String(msg.Name)
		}
		return openai.ChatCompletionMessageParamUnion{OfSystem: &system}, nil
	case "user":
		var user openai.ChatCompletionUserMessageParam
		if msg.Content != "" {
			user.Content.OfString = openai.String(msg.Content)
		}
		if msg.Name != "" {
			user.Name = openai.String(msg.Name)
		}
		return openai.ChatCompletionMessageParamUnion{OfUser: &user}, nil
	case "assistant":
		var assistant openai.ChatCompletionAssistantMessageParam
		if msg.Content != "" {
			assistant.Content.OfString = openai.String(msg.Content)
		}
		if msg.Name != "" {
			assistant.Name = openai.String(msg.Name)
		}
		if len(msg.ToolCalls) > 0 {
			toolCalls, err := toOpenAIToolCalls(msg.ToolCalls)
			if err != nil {
				return openai.ChatCompletionMessageParamUnion{}, err
			}
			assistant.ToolCalls = toolCalls
		}
		return openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant}, nil
	case "tool":
		if msg.ToolCallID == "" {
			return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("tool message missing tool_call_id")
		}
		var tool openai.ChatCompletionToolMessageParam
		if msg.Content != "" {
			tool.Content.OfString = openai.String(msg.Content)
		}
		tool.ToolCallID = msg.ToolCallID
		return openai.ChatCompletionMessageParamUnion{OfTool: &tool}, nil
	default:
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("unsupported message role: %s", msg.Role)
	}
}

func toOpenAIToolCalls(calls []ToolCall) ([]openai.ChatCompletionMessageToolCallParam, error) {
	if len(calls) == 0 {
		return nil, nil
	}

	out := make([]openai.ChatCompletionMessageToolCallParam, 0, len(calls))
	for _, call := range calls {
		if call.ID == "" {
			return nil, fmt.Errorf("tool call missing id")
		}
		if call.Function.Name == "" {
			return nil, fmt.Errorf("tool call missing function name")
		}
		out = append(out, openai.ChatCompletionMessageToolCallParam{
			ID: call.ID,
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		})
	}

	return out, nil
}

func toOpenAITools(tools []Tool) ([]openai.ChatCompletionToolParam, error) {
	out := make([]openai.ChatCompletionToolParam, 0, len(tools))
	for _, tool := range tools {
		if tool.Type != "" && tool.Type != "function" {
			return nil, fmt.Errorf("unsupported tool type: %s", tool.Type)
		}
		if tool.Function.Name == "" {
			return nil, fmt.Errorf("tool function name is required")
		}

		definition := openai.FunctionDefinitionParam{
			Name: tool.Function.Name,
		}
		if tool.Function.Description != "" {
			definition.Description = openai.String(tool.Function.Description)
		}
		if len(tool.Function.Parameters) > 0 {
			var params openai.FunctionParameters
			if err := json.Unmarshal(tool.Function.Parameters, &params); err != nil {
				return nil, fmt.Errorf("tool %s parameters: %w", tool.Function.Name, err)
			}
			definition.Parameters = params
		}

		out = append(out, openai.ChatCompletionToolParam{
			Function: definition,
		})
	}

	return out, nil
}

func fromOpenAIChatResponse(resp *openai.ChatCompletion) *ChatResponse {
	if resp == nil {
		return &ChatResponse{}
	}

	choices := make([]struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	}, 0, len(resp.Choices))

	for _, choice := range resp.Choices {
		msg := ChatMessage{
			Role:      string(choice.Message.Role),
			Content:   choice.Message.Content,
			ToolCalls: fromOpenAIToolCalls(choice.Message.ToolCalls),
		}
		choices = append(choices, struct {
			Message      ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{
			Message:      msg,
			FinishReason: choice.FinishReason,
		})
	}

	return &ChatResponse{Choices: choices}
}

func fromOpenAIToolCalls(calls []openai.ChatCompletionMessageToolCall) []ToolCall {
	if len(calls) == 0 {
		return nil
	}

	out := make([]ToolCall, 0, len(calls))
	for _, call := range calls {
		out = append(out, ToolCall{
			ID:   call.ID,
			Type: string(call.Type),
			Function: ToolCallFunc{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		})
	}

	return out
}
