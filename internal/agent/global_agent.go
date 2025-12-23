package agent

import (
	"context"
	"encoding/json"
	"time"

	"assistant-qisumi/internal/llm"
)

type GlobalAgent struct {
	llmClient              llm.Client
	chatCompletionsHandler *ChatCompletionsHandler
}

func NewGlobalAgent(llmClient llm.Client, chatCompletionsHandler *ChatCompletionsHandler) *GlobalAgent {
	return &GlobalAgent{
		llmClient:              llmClient,
		chatCompletionsHandler: chatCompletionsHandler,
	}
}

func (a *GlobalAgent) Name() string { return "global" }

func (a *GlobalAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 构造 messages + tools 调用 LLM
	messages := []llm.Message{
		{
			Role:    "system",
			Content: GlobalSystemPrompt,
		},
	}

	// 添加当前时间信息
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: "当前时间 now: " + req.Now.Format(time.RFC3339),
	})

	// 添加任务数据到系统消息
	if len(req.Tasks) > 0 {
		tasksJSON, err := json.Marshal(req.Tasks)
		if err == nil {
			messages = append(messages, llm.Message{
				Role:    "system",
				Content: "用户的任务数据（JSON格式）：\n" + string(tasksJSON),
			})
		}
	} else {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "用户目前没有任务数据。",
		})
	}

	// 添加历史消息
	messages = append(messages, historyToLLMMessages(req.Messages)...)

	// 添加当前用户输入
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: req.UserInput,
	})

	// 定义可用工具
	tools := llm.GlobalTools()

	// 使用 ChatCompletionsHandler 处理完整的工具调用流程
	// 这会自动处理：初始LLM调用 -> 工具执行 -> 二次LLM调用生成最终回复
	assistantMessage, taskPatches, err := a.chatCompletionsHandler.HandleChatCompletions(
		context.Background(),
		req.LLMConfig,
		messages,
		tools,
	)
	if err != nil {
		return nil, err
	}

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
