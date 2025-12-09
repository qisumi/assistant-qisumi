package agent

import (
	"context"
	"encoding/json"
	"time"

	"assistant-qisumi/internal/llm"
)

type ExecutorAgent struct {
	llmClient llm.Client
}

func NewExecutorAgent(llmClient llm.Client) *ExecutorAgent {
	return &ExecutorAgent{llmClient: llmClient}
}

func (a *ExecutorAgent) Name() string { return "executor" }

func (a *ExecutorAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 构造 messages + tools 调用 LLM
	messages := []llm.Message{
		{
			Role: "system",
			Content: `你是一个执行跟踪助手（Executor Agent）。

你收到的是：
- 当前任务的结构化信息（task 和 steps），已经由系统消息提供
- 最近几轮与用户的对话历史
- 用户的最新输入（user message）

你的职责：
1. 理解用户与当前任务相关的「执行类」意图，例如：
   - 标记某个步骤已完成 / 未完成 / 进行中 / 阻塞
   - 标记整个任务为完成
   - 修改任务的截止时间、优先级、标题或描述
   - 对某一步进行轻微的补充说明（detail）
2. 如果需要修改数据，必须调用提供的工具（tools）来更新任务和步骤：
   - update_task
   - update_steps
3. 修改完成后，你需要给用户一段自然语言回复，说明你做了哪些更新，以及下一步建议。

重要原则：
- 所有对任务 / 步骤的修改必须通过工具调用完成，不能只在文字里说“我已经修改了”。
- 只使用系统消息里给出的 task_id 和 step_id，不要编造不存在的 ID。
- 如果用户说的是模糊描述（例如“我完成了读论文那一步”），你需要通过标题/detail 内容模糊匹配，找到最合适的 step_id，如果无法确定，就向用户确认。
- 如果用户只是提出疑问（例如“这个任务还有几步没完成？”），你可以只给出自然语言回答，不调用任何工具。

输出流程（对你来说）：
1. 如果需要修改：先调用工具（如 update_steps），再在工具调用返回之后给出最终的自然语言回答。
2. 如果不需要修改：直接给出自然语言回答即可。`,
		},
	}

	// 添加当前任务状态信息
	if req.Task != nil {
		if taskJSON, err := json.Marshal(req.Task); err == nil {
			messages = append(messages, llm.Message{
				Role:    "system",
				Content: "当前任务状态（只读 JSON）：\n" + string(taskJSON),
			})
		}
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
