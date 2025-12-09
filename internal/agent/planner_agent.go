package agent

import (
	"assistant-qisumi/internal/llm"
	"context"
	"encoding/json"
	"fmt"
)

type PlannerAgent struct {
	llmClient llm.Client
}

func NewPlannerAgent(llmClient llm.Client) *PlannerAgent {
	return &PlannerAgent{llmClient: llmClient}
}

func (a *PlannerAgent) Name() string { return "planner" }

func (a *PlannerAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 构造系统提示词
	systemPrompt := `你是一个任务规划助手（Planner Agent）。

你收到的是：
- 当前任务的结构化信息（task 和 steps）
- 当前时间 now
- 用户的最新请求，通常包含“重新规划”“拆解”“重排”“延期后重新安排”等意图

你的职责：
1. 针对当前任务进行结构性的调整，包括但不限于：
   - 重新拆分某一步为多个更细的子步骤
   - 新增或删除步骤
   - 调整步骤的执行顺序（order_index）
   - 根据任务的新截止时间 / 当前进度，重新规划步骤的 planned_start / planned_end
   - 按用户描述创建任务依赖关系（例如“任务A完成后再开始任务B的第一步”）
2. 所有结构性变更必须通过 tools 实现：
   - add_steps：新增步骤或子步骤
   - update_steps：修改步骤标题、描述、顺序、估时、状态、计划时间
   - update_task：更新任务的整体信息（如 due_at、priority）
   - add_dependencies：在任务或步骤之间创建依赖关系
3. 合理地使用 estimate_minutes 和时间窗口：
   - 如果用户给出了明确时间，你要尽量遵循
   - 如果用户只给出“这周内”“今晚”等模糊描述，你可以根据 now 和 due_at 进行合理推断

输出要求：
1. 优先确保工具调用正确、参数齐全，不要出现多余字段。
2. 工具调用完成后，用一段简洁自然语言告诉用户：
   - 新的步骤结构是什么（可以简要列出）
   - 大致执行顺序和时间安排
   - 如果有依赖关系，也要提一下「某任务完成后会自动解锁 XXX 步骤」。

安全与约束：
- 只修改与当前任务真正相关的内容，不要随意创建额外任务。
- 对于非常模糊的描述，如果你不确定具体怎么拆分，先做一版合理的初稿，并在自然语言回复里提醒用户可以继续调整。`

	// 2. 构造任务上下文
	baseContext := `当前任务状态（JSON，只读）：
{
  "task_id": %d,
  "title": "%s",
  "description": "%s",
  "status": "%s",
  "priority": "%s",
  "due_at": "%s",
  "steps": [
`

	taskContext := fmt.Sprintf(baseContext, req.Task.ID, req.Task.Title, req.Task.Description, req.Task.Status, req.Task.Priority, req.Task.DueAt)

	for _, step := range req.Task.Steps {
		taskContext += fmt.Sprintf(`    {
      "step_id": %d,
      "title": "%s",
      "detail": "%s",
      "status": "%s",
      "estimate_minutes": %d,
      "order_index": %d
    },
`, step.ID, step.Title, step.Detail, step.Status, step.EstimateMin, step.OrderIndex)
	}

	// 移除最后一个逗号
	if len(req.Task.Steps) > 0 {
		taskContext = taskContext[:len(taskContext)-2] + "\n"
	}

	taskContext += fmt.Sprintf(`  ]
}

当前时间 now: %s`, req.Now.Format("2006-01-02T15:04:05"))

	// 3. 准备tools
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
					"required":             []string{"task_id", "fields"},
					"additionalProperties": false,
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
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "add_steps",
				Description: "Add new steps to an existing task.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{
							"type": "integer",
						},
						"parent_step_id": map[string]interface{}{
							"type":        []string{"integer", "null"},
							"description": "Optional parent step ID for substeps. Use null for top-level steps.",
						},
						"steps": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"title":            map[string]interface{}{"type": "string"},
									"detail":           map[string]interface{}{"type": "string"},
									"estimate_minutes": map[string]interface{}{"type": "integer", "minimum": 1},
									"insert_after_step_id": map[string]interface{}{
										"type":        []string{"integer", "null"},
										"description": "Insert after this step. If null, append to the end.",
									},
								},
								"required":             []string{"title"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"task_id", "steps"},
					"additionalProperties": false,
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "add_dependencies",
				Description: "Create dependencies between tasks or steps. When the predecessor is done, the successor can be unlocked or activated.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"items": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"predecessor_task_id": map[string]interface{}{
										"type": "integer",
									},
									"predecessor_step_id": map[string]interface{}{
										"type":        []string{"integer", "null"},
										"description": "Optional step ID. If null, the whole task is the predecessor.",
									},
									"successor_task_id": map[string]interface{}{
										"type": "integer",
									},
									"successor_step_id": map[string]interface{}{
										"type":        []string{"integer", "null"},
										"description": "Optional step ID. If null, the whole task is the successor.",
									},
									"condition": map[string]interface{}{
										"type": "string",
										"enum": []string{"task_done", "step_done"},
									},
									"action": map[string]interface{}{
										"type": "string",
										"enum": []string{"unlock_step", "set_task_todo", "notify_only"},
									},
								},
								"required":             []string{"predecessor_task_id", "successor_task_id", "condition", "action"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"items"},
					"additionalProperties": false,
				},
			},
		},
	}

	// 4. 准备messages
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "system", Content: taskContext},
		{Role: "user", Content: req.UserInput},
	}

	// 5. 调用LLM
	chatReq := llm.ChatRequest{
		Model:      req.LLMConfig.Model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
	}

	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// 6. 处理LLM响应
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	chatMsg := resp.Choices[0].Message
	assistantMessage := chatMsg.Content
	var taskPatches []TaskPatch

	// 7. 处理工具调用
	for _, toolCall := range chatMsg.ToolCalls {
		// 解析工具调用参数
		var toolArgs map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &toolArgs); err != nil {
			continue
		}

		// 根据工具名称生成对应的TaskPatch
		switch toolCall.Function.Name {
		case "update_task":
			taskPatches = append(taskPatches, TaskPatch{
				Type:    "UpdateTask",
				Payload: toolArgs,
			})
		case "update_steps":
			taskPatches = append(taskPatches, TaskPatch{
				Type:    "UpdateSteps",
				Payload: toolArgs,
			})
		case "add_steps":
			taskPatches = append(taskPatches, TaskPatch{
				Type:    "AddSteps",
				Payload: toolArgs,
			})
		case "add_dependencies":
			taskPatches = append(taskPatches, TaskPatch{
				Type:    "AddDependencies",
				Payload: toolArgs,
			})
		}
	}

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
