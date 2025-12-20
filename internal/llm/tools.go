package llm

import "encoding/json"

// MustRawJSON 帮助函数：把字符串转为 json.RawMessage（不做错误处理，初始化阶段 panic 重启即可）
func MustRawJSON(s string) json.RawMessage {
	return json.RawMessage(s)
}

// CommonTools 返回 Executor / Planner / Global 等通用会用到的一组工具。
func CommonTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "update_task",
				Description: "Update a task's metadata such as title, description, status, priority or due_at.",
				Parameters: MustRawJSON(`{
          "type": "object",
          "properties": {
            "task_id": { "type": "integer", "description": "The ID of the task to update." },
            "fields": {
              "type": "object",
              "description": "Fields to update. Only include fields that need to be changed.",
              "properties": {
                "title":        { "type": "string" },
                "description":  { "type": "string" },
                "status": {
                  "type": "string",
                  "enum": ["todo", "in_progress", "done", "cancelled"]
                },
                "priority": {
                  "type": "string",
                  "enum": ["low", "medium", "high"]
                },
                "due_at": {
                  "type": "string",
                  "description": "New due date time in ISO 8601 format, e.g. 2025-12-08T20:00:00"
                }
              },
              "additionalProperties": false
            }
          },
          "required": ["task_id", "fields"],
          "additionalProperties": false
        }`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "update_steps",
				Description: "Update one or more existing steps in a task.",
				Parameters: MustRawJSON(`{
          "type": "object",
          "properties": {
            "task_id": { "type": "integer" },
            "updates": {
              "type": "array",
              "items": {
                "type": "object",
                "properties": {
                  "step_id": { "type": "integer" },
                  "fields": {
                    "type": "object",
                    "properties": {
                      "title":    { "type": "string" },
                      "detail":   { "type": "string" },
                      "status": {
                        "type": "string",
                        "enum": ["locked", "todo", "in_progress", "done", "blocked"]
                      },
                      "blocking_reason": { "type": "string" },
                      "estimate_minutes": {
                        "type": "integer",
                        "minimum": 1
                      },
                      "order_index": {
                        "type": "integer",
                        "description": "New order index, smaller means earlier."
                      },
                      "planned_start": {
                        "type": "string",
                        "description": "Planned start time in ISO 8601."
                      },
                      "planned_end": {
                        "type": "string",
                        "description": "Planned end time in ISO 8601."
                      }
                    },
                    "additionalProperties": false
                  }
                },
                "required": ["step_id", "fields"],
                "additionalProperties": false
              }
            }
          },
          "required": ["task_id", "updates"],
          "additionalProperties": false
        }`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "add_steps",
				Description: "Add new steps to an existing task.",
				Parameters: MustRawJSON(`{
          "type": "object",
          "properties": {
            "task_id": { "type": "integer" },
            "parent_step_id": {
              "type": ["integer", "null"],
              "description": "Optional parent step ID for substeps. Use null for top-level steps."
            },
            "steps": {
              "type": "array",
              "items": {
                "type": "object",
                "properties": {
                  "title":  { "type": "string" },
                  "detail": { "type": "string" },
                  "estimate_minutes": {
                    "type": "integer",
                    "minimum": 1
                  },
                  "insert_after_step_id": {
                    "type": ["integer", "null"],
                    "description": "Insert after this step. If null, append to the end."
                  }
                },
                "required": ["title"],
                "additionalProperties": false
              }
            }
          },
          "required": ["task_id", "steps"],
          "additionalProperties": false
        }`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "add_dependencies",
				Description: "Create dependencies between tasks or steps. When the predecessor is done, the successor can be unlocked or activated.",
				Parameters: MustRawJSON(`{
          "type": "object",
          "properties": {
            "items": {
              "type": "array",
              "items": {
                "type": "object",
                "properties": {
                  "predecessor_task_id": { "type": "integer" },
                  "predecessor_step_id": {
                    "type": ["integer", "null"],
                    "description": "Optional step ID. If null, the whole task is the predecessor."
                  },
                  "successor_task_id": { "type": "integer" },
                  "successor_step_id": {
                    "type": ["integer", "null"],
                    "description": "Optional step ID. If null, the whole task is the successor."
                  },
                  "condition": {
                    "type": "string",
                    "enum": ["task_done", "step_done"]
                  },
                  "action": {
                    "type": "string",
                    "enum": ["unlock_step", "set_task_todo", "notify_only"]
                  }
                },
                "required": ["predecessor_task_id", "successor_task_id", "condition", "action"],
                "additionalProperties": false
              }
            }
          },
          "required": ["items"],
          "additionalProperties": false
        }`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "mark_tasks_focus_today",
				Description: "Mark one or more tasks as today's focus tasks, for daily planning or overview.",
				Parameters: MustRawJSON(`{
          "type": "object",
          "properties": {
            "task_ids": {
              "type": "array",
              "items": { "type": "integer" }
            }
          },
          "required": ["task_ids"],
          "additionalProperties": false
        }`),
			},
		},
	}
}

func ExecutorTools() []Tool {
	return []Tool{
		CommonTools()[0], // update_task
		CommonTools()[1], // update_steps
	}
}

func PlannerTools() []Tool {
	tools := CommonTools()
	// update_task, update_steps, add_steps, add_dependencies
	return tools[:4]
}

func GlobalTools() []Tool {
	return []Tool{
		CommonTools()[0], // update_task
		CommonTools()[4], // mark_tasks_focus_today
	}
}
