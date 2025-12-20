package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/task"
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
	ctx context.Context,
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

	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := h.llmClient.Chat(ctx, cfg, chatReq)
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
			patches, err := h.generateTaskPatchesFromToolCall(toolCall)
			if err != nil {
				return "", nil, fmt.Errorf("failed to generate task patches from tool call %s: %w", toolCall.Function.Name, err)
			}
			taskPatches = append(taskPatches, patches...)
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

			finalResp, err := h.llmClient.Chat(ctx, cfg, finalChatReq)
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

// generateTaskPatchesFromToolCall 从工具调用生成TaskPatch列表
func (h *ChatCompletionsHandler) generateTaskPatchesFromToolCall(toolCall llm.ToolCall) ([]TaskPatch, error) {
	var patches []TaskPatch
	name := toolCall.Function.Name
	argsJSON := toolCall.Function.Arguments

	switch name {
	case "update_task":
		var args UpdateTaskArgs
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("update_task args decode: %w", err)
		}
		patches = append(patches, TaskPatch{
			Kind: PatchUpdateTask,
			UpdateTask: &UpdateTaskPatch{
				TaskID: args.TaskID,
				Fields: args.Fields,
			},
		})

	case "update_steps":
		var args UpdateStepsArgs
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("update_steps args decode: %w", err)
		}
		for _, u := range args.Updates {
			patches = append(patches, TaskPatch{
				Kind: PatchUpdateStep,
				UpdateStep: &UpdateStepPatch{
					TaskID: args.TaskID,
					StepID: u.StepID,
					Fields: u.Fields,
				},
			})
		}

	case "add_steps":
		var args AddStepsArgs
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("add_steps args decode: %w", err)
		}
		records := make([]task.NewStepRecord, 0, len(args.Steps))
		for _, s := range args.Steps {
			records = append(records, task.NewStepRecord{
				Title:             s.Title,
				Detail:            s.Detail,
				EstimateMinutes:   s.EstimateMinutes,
				InsertAfterStepID: s.InsertAfterStepID,
			})
		}
		patches = append(patches, TaskPatch{
			Kind: PatchAddSteps,
			AddSteps: &AddStepsPatch{
				TaskID:        args.TaskID,
				ParentStepID:  args.ParentStepID,
				StepsToInsert: records,
			},
		})

	case "add_dependencies":
		var args AddDependenciesArgs
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("add_dependencies args decode: %w", err)
		}
		items := make([]task.DependencyItem, 0, len(args.Items))
		for _, it := range args.Items {
			items = append(items, task.DependencyItem{
				PredecessorTaskID: it.PredecessorTaskID,
				PredecessorStepID: it.PredecessorStepID,
				SuccessorTaskID:   it.SuccessorTaskID,
				SuccessorStepID:   it.SuccessorStepID,
				Condition:         it.Condition,
				Action:            it.Action,
			})
		}
		patches = append(patches, TaskPatch{
			Kind: PatchAddDependencies,
			AddDependencies: &AddDependenciesPatch{
				Items: items,
			},
		})

	case "mark_tasks_focus_today":
		var args MarkTasksFocusTodayArgs
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("mark_tasks_focus_today args decode: %w", err)
		}
		patches = append(patches, TaskPatch{
			Kind: PatchMarkTasksFocusToday,
			MarkTasksFocusToday: &MarkTasksFocusTodayPatch{
				TaskIDs: args.TaskIDs,
			},
		})
	}

	return patches, nil
}
