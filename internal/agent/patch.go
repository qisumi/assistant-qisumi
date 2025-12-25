package agent

import "assistant-qisumi/internal/domain"

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

	UpdateTask          *UpdateTaskPatch          `json:"updateTask,omitempty"`
	UpdateStep          *UpdateStepPatch          `json:"updateStep,omitempty"`
	AddSteps            *AddStepsPatch            `json:"addSteps,omitempty"`
	AddDependencies     *AddDependenciesPatch     `json:"addDependencies,omitempty"`
	MarkTasksFocusToday *MarkTasksFocusTodayPatch `json:"markTasksFocusToday,omitempty"`
	CreateTask          *CreateTaskPatch          `json:"createTask,omitempty"`
}

// --- 各种具体 Patch Payload ---

type UpdateTaskPatch struct {
	TaskID uint64                  `json:"taskId"`
	Fields domain.UpdateTaskFields `json:"fields"`
}

type UpdateStepPatch struct {
	TaskID uint64                  `json:"taskId"`
	StepID uint64                  `json:"stepId"`
	Fields domain.UpdateStepFields `json:"fields"`
}

type AddStepsPatch struct {
	TaskID        uint64                 `json:"taskId"`
	ParentStepID  *uint64                `json:"parentStepId,omitempty"`
	StepsToInsert []domain.NewStepRecord `json:"stepsToInsert"`
}

type AddDependenciesPatch struct {
	Items []domain.DependencyItem `json:"items"`
}

type MarkTasksFocusTodayPatch struct {
	TaskIDs []uint64 `json:"taskIds"`
}

type CreateTaskPatch struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	DueAt       *string                `json:"dueAt,omitempty"`
	Priority    string                 `json:"priority"`
	Steps       []domain.NewStepRecord `json:"steps"`
}
