package agent

import "assistant-qisumi/internal/task"

// 对应 tool: update_task
type UpdateTaskArgs struct {
	TaskID uint64                `json:"task_id"`
	Fields task.UpdateTaskFields `json:"fields"`
}

// 对应 tool: update_steps
type UpdateStepsArgs struct {
	TaskID  uint64               `json:"task_id"`
	Updates []UpdateStepItemArgs `json:"updates"`
}

type UpdateStepItemArgs struct {
	StepID uint64                `json:"step_id"`
	Fields task.UpdateStepFields `json:"fields"`
}

// 对应 tool: add_steps
type AddStepsArgs struct {
	TaskID       uint64             `json:"task_id"`
	ParentStepID *uint64            `json:"parent_step_id"`
	Steps        []NewStepArgsInput `json:"steps"`
}

type NewStepArgsInput struct {
	Title             string  `json:"title"`
	Detail            string  `json:"detail"`
	EstimateMinutes   *int    `json:"estimate_minutes"`
	InsertAfterStepID *uint64 `json:"insert_after_step_id"`
}

// 对应 tool: add_dependencies
type AddDependenciesArgs struct {
	Items []DependencyItemArgs `json:"items"`
}

// Note: DependencyItemArgs is still here because it's slightly different from DependencyItem (pointers for optional fields)
// Actually, let's check if we can use task.DependencyItem directly or if we need a separate one for tool args.
// The tool args usually match the JSON schema of the tool.

type DependencyItemArgs struct {
	PredecessorTaskID uint64  `json:"predecessor_task_id"`
	PredecessorStepID *uint64 `json:"predecessor_step_id"`
	SuccessorTaskID   uint64  `json:"successor_task_id"`
	SuccessorStepID   *uint64 `json:"successor_step_id"`
	Condition         string  `json:"condition"`
	Action            string  `json:"action"`
}

// 对应 tool: mark_tasks_focus_today
type MarkTasksFocusTodayArgs struct {
	TaskIDs []uint64 `json:"task_ids"`
}
