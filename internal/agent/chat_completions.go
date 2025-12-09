package agent

import (
	"assistant-qisumi/internal/llm"
	"encoding/json"
	"fmt"
)

// ChatCompletionsHandler 处理完整的Chat Completions流程，包括工具调用和结果处理
type ChatCompletionsHandler struct {
	llmClient llm.Client
	toolMap   map[string]ToolExecutor
}

// ToolExecutor 工具执行器接口
type ToolExecutor interface {
	Execute(args string) (interface{}, error)
}

// NewChatCompletionsHandler 创建Chat Completions处理器
func NewChatCompletionsHandler(llmClient llm.Client, toolMap map[string]ToolExecutor) *ChatCompletionsHandler {
	return &ChatCompletionsHandler{
		llmClient: llmClient,
		toolMap:   toolMap,
	}
}

// HandleChatCompletions 处理完整的Chat Completions流程
func (h *ChatCompletionsHandler) HandleChatCompletions(
	cfg llm.Config,
	initialMessages []llm.Message,
	tools []llm.Tool,
) (string, []TaskPatch, error) {
	// 1. 初始LLM调用
	chatReq := llm.ChatRequest{
		Model:      cfg.Model,
		Messages:   initialMessages,
		Tools:      tools,
		ToolChoice: "auto",
	}

	resp, err := h.llmClient.Chat(nil, cfg, chatReq)
	if err != nil {
		return "", nil, err
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in llm response")
	}

	choice := resp.Choices[0]
	assistantMessage := choice.Message.Content

	// 2. 检查是否需要工具调用
	if len(choice.Message.ToolCalls) > 0 {
		// 3. 处理工具调用
		var taskPatches []TaskPatch
		var toolResponses []llm.Message

		for _, toolCall := range choice.Message.ToolCalls {
			// 执行工具调用
			toolResp, err := h.executeToolCall(toolCall)
			if err != nil {
				return "", nil, fmt.Errorf("failed to execute tool %s: %w", toolCall.Function.Name, err)
			}

			// 生成工具响应消息
			toolRespMsg := llm.Message{
				Role:       "tool",
				Content:    string(toolResp),
				ToolCallID: toolCall.ID,
				Name:       toolCall.Function.Name,
			}
			toolResponses = append(toolResponses, toolRespMsg)

			// 解析工具调用结果生成TaskPatch
			patch, err := h.generateTaskPatchFromToolCall(toolCall, toolResp)
			if err != nil {
				return "", nil, fmt.Errorf("failed to generate task patch from tool call %s: %w", toolCall.Function.Name, err)
			}
			if patch != nil {
				taskPatches = append(taskPatches, *patch)
			}
		}

		// 4. 二次LLM调用，生成最终回复
		if len(toolResponses) > 0 {
			// 合并消息
			finalMessages := append(initialMessages, llm.Message{
				Role:      "assistant",
				Content:   assistantMessage,
				ToolCalls: choice.Message.ToolCalls,
			})
			finalMessages = append(finalMessages, toolResponses...)

			// 调用LLM生成最终回复
			finalChatReq := llm.ChatRequest{
				Model:      cfg.Model,
				Messages:   finalMessages,
				Tools:      tools,
				ToolChoice: "none", // 不再调用工具，直接生成回复
			}

			finalResp, err := h.llmClient.Chat(nil, cfg, finalChatReq)
			if err != nil {
				return "", nil, err
			}

			if len(finalResp.Choices) > 0 {
				assistantMessage = finalResp.Choices[0].Message.Content
			}
		}

		return assistantMessage, taskPatches, nil
	}

	// 不需要工具调用，直接返回LLM回复
	return assistantMessage, nil, nil
}

// executeToolCall 执行单个工具调用
func (h *ChatCompletionsHandler) executeToolCall(toolCall llm.ToolCall) ([]byte, error) {
	executor, ok := h.toolMap[toolCall.Function.Name]
	if !ok {
		return nil, fmt.Errorf("tool executor not found for %s", toolCall.Function.Name)
	}

	result, err := executor.Execute(toolCall.Function.Arguments)
	if err != nil {
		return nil, err
	}

	// 序列化工具执行结果
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool result: %w", err)
	}

	return resultJSON, nil
}

// generateTaskPatchFromToolCall 从工具调用生成TaskPatch
func (h *ChatCompletionsHandler) generateTaskPatchFromToolCall(toolCall llm.ToolCall, toolResp []byte) (*TaskPatch, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return nil, err
	}

	switch toolCall.Function.Name {
	case "update_task":
		if taskID, ok := args["task_id"].(float64); ok {
			if fields, ok := args["fields"].(map[string]interface{}); ok {
				payload := map[string]interface{}{
					"task_id": uint64(taskID),
					"fields":  fields,
				}
				return &TaskPatch{
					Type:    "update_task",
					Payload: payload,
				}, nil
			}
		}
	case "update_steps":
		if taskID, ok := args["task_id"].(float64); ok {
			if updates, ok := args["updates"].([]interface{}); ok {
				for _, update := range updates {
					if updateMap, ok := update.(map[string]interface{}); ok {
						if stepID, ok := updateMap["step_id"].(float64); ok {
							if fields, ok := updateMap["fields"].(map[string]interface{}); ok {
								payload := map[string]interface{}{
									"task_id": uint64(taskID),
									"step_id": uint64(stepID),
									"fields":  fields,
								}
								return &TaskPatch{
									Type:    "update_step",
									Payload: payload,
								}, nil
							}
						}
					}
				}
			}
		}
	case "add_steps":
		if taskID, ok := args["task_id"].(float64); ok {
			return &TaskPatch{
				Type: "add_steps",
				Payload: map[string]interface{}{
					"task_id":   uint64(taskID),
					"arguments": string(toolCall.Function.Arguments),
				},
			}, nil
		}
	case "add_dependencies":
		return &TaskPatch{
			Type: "add_dependencies",
			Payload: map[string]interface{}{
				"arguments": string(toolCall.Function.Arguments),
			},
		}, nil
	case "mark_tasks_focus_today":
		return &TaskPatch{
			Type: "mark_tasks_focus_today",
			Payload: map[string]interface{}{
				"arguments": string(toolCall.Function.Arguments),
			},
		}, nil
	}

	return nil, nil
}
