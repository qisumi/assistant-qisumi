package agent

import (
	"context"
	"encoding/json"
	"time"

	"assistant-qisumi/internal/llm"
)

type GlobalAgent struct {
	llmClient llm.Client
}

func NewGlobalAgent(llmClient llm.Client) *GlobalAgent {
	return &GlobalAgent{llmClient: llmClient}
}

func (a *GlobalAgent) Name() string { return "global" }

func (a *GlobalAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 构造 messages + tools 调用 LLM
	messages := []llm.Message{
		{
			Role: "system",
			Content: `你是一个跨任务日程规划助手（Global Agent）。

你收到的是：
- 用户今天/本周的任务概览（由系统消息提供，包含多个 task 列表）
  - 每个任务包含：title, status, priority, due_at
  - 每个任务下有若干关键步骤（title, status, estimate_minutes, planned_start/planned_end）
- 用户的提问，例如：
  - 「我今天要做什么？」
  - 「帮我看一下这周的安排」
  - 「有没有已经过期但没完成的任务？」

你的职责：
1. 基于给出的任务数据，为用户生成一个清晰的「计划说明」，例如：
   - 今日待办清单（按优先级和紧迫度排序）
   - 已过期但未完成的任务提醒
   - 建议的执行顺序（可以按时间或能量水平来安排）
2. 你可以使用工具：
   - mark_tasks_focus_today：标记今天重点关注的任务（如果用户有此意图）
   - update_task / update_steps：仅在用户明确要求修改时使用（例如「帮我把某任务优先级调高」）
3. 输出中尽量包含结构化层次：
   - 第一部分：今日重点任务
   - 第二部分：可选任务/轻量任务
   - 第三部分：过期任务的提醒

注意：
- 你不会自己查询数据库，你只能使用系统给你的任务数据。
- 不要随意修改任务状态，除非用户有明确指令。`,
		},
	}

	// 添加当前时间信息
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: "当前时间 now: " + req.Now.Format(time.RFC3339),
	})

	// 添加历史消息
	for _, msg := range req.Messages {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前用户输入
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: req.UserInput,
	})

	// 定义可用工具
	tools := []llm.Tool{
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "mark_tasks_focus_today",
				Description: "Mark one or more tasks as today's focus tasks, for daily planning or overview.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_ids": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"type": "integer"},
						},
					},
					"required":             []string{"task_ids"},
					"additionalProperties": false,
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "update_task",
				Description: "Update a task's metadata such as title, description, status, priority or due_at.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{
							"type":        "integer",
							"description": "The ID of the task to update.",
						},
						"fields": map[string]interface{}{
							"type":        "object",
							"description": "Fields to update. Only include fields that need to be changed.",
							"properties": map[string]interface{}{
								"title":       map[string]interface{}{"type": "string"},
								"description": map[string]interface{}{"type": "string"},
								"status": map[string]interface{}{
									"type": "string",
									"enum": []string{"todo", "in_progress", "done", "cancelled"},
								},
								"priority": map[string]interface{}{
									"type": "string",
									"enum": []string{"low", "medium", "high"},
								},
								"due_at": map[string]interface{}{
									"type":        "string",
									"description": "New due date time in ISO 8601 format, e.g. 2025-12-08T20:00:00",
								},
							},
							"additionalProperties": false,
						},
					},
					"required": []string{"task_id", "fields"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "update_steps",
				Description: "Update one or more existing steps in a task.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{
							"type": "integer",
						},
						"updates": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"step_id": map[string]interface{}{
										"type": "integer",
									},
									"fields": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"title":  map[string]interface{}{"type": "string"},
											"detail": map[string]interface{}{"type": "string"},
											"status": map[string]interface{}{
												"type": "string",
												"enum": []string{"locked", "todo", "in_progress", "done", "blocked"},
											},
											"blocking_reason":  map[string]interface{}{"type": "string"},
											"estimate_minutes": map[string]interface{}{"type": "integer", "minimum": 1},
											"order_index": map[string]interface{}{
												"type":        "integer",
												"description": "New order index, smaller means earlier.",
											},
											"planned_start": map[string]interface{}{
												"type":        "string",
												"description": "Planned start time in ISO 8601.",
											},
											"planned_end": map[string]interface{}{
												"type":        "string",
												"description": "Planned end time in ISO 8601.",
											},
										},
										"additionalProperties": false,
									},
								},
								"required":             []string{"step_id", "fields"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"task_id", "updates"},
					"additionalProperties": false,
				},
			},
		},
	}

	// 构造Chat请求
	chatReq := llm.ChatRequest{
		Model:      req.LLMConfig.Model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
	}

	// 调用LLM
	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		return nil, err
	}

	// 2. 处理 LLM 响应，生成 assistant 回复文本和 TaskPatches
	var assistantMessage string
	var taskPatches []TaskPatch

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		assistantMessage = choice.Message.Content

		// 解析工具调用
		for _, toolCall := range choice.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				continue
			}

			switch toolCall.Function.Name {
			case "mark_tasks_focus_today":
				if taskIDs, ok := args["task_ids"].([]interface{}); ok {
					var uintTaskIDs []uint64
					for _, id := range taskIDs {
						if idFloat, ok := id.(float64); ok {
							uintTaskIDs = append(uintTaskIDs, uint64(idFloat))
						}
					}
					payload := map[string]interface{}{
						"task_ids": uintTaskIDs,
					}
					taskPatches = append(taskPatches, TaskPatch{
						Type:    "mark_tasks_focus_today",
						Payload: payload,
					})
				}
			case "update_task":
				if taskID, ok := args["task_id"].(float64); ok {
					if fields, ok := args["fields"].(map[string]interface{}); ok {
						payload := map[string]interface{}{
							"task_id": uint64(taskID),
							"fields":  fields,
						}
						taskPatches = append(taskPatches, TaskPatch{
							Type:    "update_task",
							Payload: payload,
						})
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
										taskPatches = append(taskPatches, TaskPatch{
											Type:    "update_step",
											Payload: payload,
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
