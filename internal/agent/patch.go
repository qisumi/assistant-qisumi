package agent

import "assistant-qisumi/internal/task"

type PatchKind string

const (
	PatchUpdateTask          PatchKind = "update_task"
	PatchUpdateStep          PatchKind = "update_step"
	PatchAddSteps            PatchKind = "add_steps"
	PatchAddDependencies     PatchKind = "add_dependencies"
	PatchMarkTasksFocusToday PatchKind = "mark_tasks_focus_today"
	PatchCreateTask          PatchKind = "create_task"
)

// 顶层 Patch，Kind 决定哪个字段非 nil
type TaskPatch struct {
	Kind PatchKind `json:"kind"`

	UpdateTask          *UpdateTaskPatch          `json:"update_task,omitempty"`
	UpdateStep          *UpdateStepPatch          `json:"update_step,omitempty"`
	AddSteps            *AddStepsPatch            `json:"add_steps,omitempty"`
	AddDependencies     *AddDependenciesPatch     `json:"add_dependencies,omitempty"`
	MarkTasksFocusToday *MarkTasksFocusTodayPatch `json:"mark_tasks_focus_today,omitempty"`
	CreateTask          *CreateTaskPatch          `json:"create_task,omitempty"`
}

// --- 各种具体 Patch Payload ---

type UpdateTaskPatch struct {
	TaskID uint64                `json:"task_id"`
	Fields task.UpdateTaskFields `json:"fields"`
}

type UpdateStepPatch struct {
	TaskID uint64                `json:"task_id"`
	StepID uint64                `json:"step_id"`
	Fields task.UpdateStepFields `json:"fields"`
}

type AddStepsPatch struct {
	TaskID        uint64               `json:"task_id"`
	ParentStepID  *uint64              `json:"parent_step_id,omitempty"`
	StepsToInsert []task.NewStepRecord `json:"steps_to_insert"`
}

type AddDependenciesPatch struct {
	Items []task.DependencyItem `json:"items"`
}

type MarkTasksFocusTodayPatch struct {
	TaskIDs []uint64 `json:"task_ids"`
}

type CreateTaskPatch struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	DueAt       *string              `json:"due_at,omitempty"`
	Priority    string               `json:"priority"`
	Steps       []task.NewStepRecord `json:"steps"`
}
