package tool

import (
	"assistant-qisumi/internal/task"
	"time"
)

// UpdateTaskArguments update_task 工具参数
type UpdateTaskArguments struct {
	TaskID uint64                 `json:"task_id"`
	Fields map[string]interface{} `json:"fields"`
}

func (args *UpdateTaskArguments) Validate() error {
	// 简单验证，确保 task_id 不为 0
	if args.TaskID == 0 {
		return ErrInvalidTaskID
	}
	return nil
}

// UpdateStepsArguments update_steps 工具参数
type UpdateStepsArguments struct {
	TaskID  uint64       `json:"task_id"`
	Updates []StepUpdate `json:"updates"`
}

// StepUpdate 单个步骤更新
type StepUpdate struct {
	StepID uint64                 `json:"step_id"`
	Fields map[string]interface{} `json:"fields"`
}

func (args *UpdateStepsArguments) Validate() error {
	if args.TaskID == 0 {
		return ErrInvalidTaskID
	}
	if len(args.Updates) == 0 {
		return ErrEmptyUpdates
	}
	for _, update := range args.Updates {
		if update.StepID == 0 {
			return ErrInvalidStepID
		}
	}
	return nil
}

// AddStepsArguments add_steps 工具参数
type AddStepsArguments struct {
	TaskID       uint64    `json:"task_id"`
	ParentStepID *uint64   `json:"parent_step_id"`
	Steps        []NewStep `json:"steps"`
}

// NewStep 新步骤
type NewStep struct {
	Title             string  `json:"title"`
	Detail            string  `json:"detail"`
	EstimateMinutes   int     `json:"estimate_minutes"`
	InsertAfterStepID *uint64 `json:"insert_after_step_id"`
}

func (args *AddStepsArguments) Validate() error {
	if args.TaskID == 0 {
		return ErrInvalidTaskID
	}
	if len(args.Steps) == 0 {
		return ErrEmptySteps
	}
	for _, step := range args.Steps {
		if step.Title == "" {
			return ErrEmptyStepTitle
		}
		if step.EstimateMinutes < 1 {
			return ErrInvalidEstimateMinutes
		}
	}
	return nil
}

// AddDependenciesArguments add_dependencies 工具参数
type AddDependenciesArguments struct {
	Items []DependencyItem `json:"items"`
}

// DependencyItem 依赖项
type DependencyItem struct {
	PredecessorTaskID uint64  `json:"predecessor_task_id"`
	PredecessorStepID *uint64 `json:"predecessor_step_id"`
	SuccessorTaskID   uint64  `json:"successor_task_id"`
	SuccessorStepID   *uint64 `json:"successor_step_id"`
	Condition         string  `json:"condition"`
	Action            string  `json:"action"`
}

func (args *AddDependenciesArguments) Validate() error {
	if len(args.Items) == 0 {
		return ErrEmptyDependencies
	}
	for _, item := range args.Items {
		if item.PredecessorTaskID == 0 {
			return ErrInvalidPredecessorTaskID
		}
		if item.SuccessorTaskID == 0 {
			return ErrInvalidSuccessorTaskID
		}
		if item.Condition != "task_done" && item.Condition != "step_done" {
			return ErrInvalidCondition
		}
		if item.Action != "unlock_step" && item.Action != "set_task_todo" && item.Action != "notify_only" {
			return ErrInvalidAction
		}
	}
	return nil
}

// MarkTasksFocusTodayArguments mark_tasks_focus_today 工具参数
type MarkTasksFocusTodayArguments struct {
	TaskIDs []uint64 `json:"task_ids"`
}

func (args *MarkTasksFocusTodayArguments) Validate() error {
	if len(args.TaskIDs) == 0 {
		return ErrEmptyTaskIDs
	}
	for _, taskID := range args.TaskIDs {
		if taskID == 0 {
			return ErrInvalidTaskID
		}
	}
	return nil
}

// 错误定义
var (
	ErrInvalidTaskID            = NewToolError("invalid task_id")
	ErrInvalidStepID            = NewToolError("invalid step_id")
	ErrEmptyUpdates             = NewToolError("no updates provided")
	ErrEmptySteps               = NewToolError("no steps provided")
	ErrEmptyStepTitle           = NewToolError("step title cannot be empty")
	ErrInvalidEstimateMinutes   = NewToolError("estimate_minutes must be at least 1")
	ErrEmptyDependencies        = NewToolError("no dependencies provided")
	ErrInvalidPredecessorTaskID = NewToolError("invalid predecessor_task_id")
	ErrInvalidSuccessorTaskID   = NewToolError("invalid successor_task_id")
	ErrInvalidCondition         = NewToolError("invalid condition, must be 'task_done' or 'step_done'")
	ErrInvalidAction            = NewToolError("invalid action, must be 'unlock_step', 'set_task_todo', or 'notify_only'")
	ErrEmptyTaskIDs             = NewToolError("no task_ids provided")
)

// ToolError 工具错误
type ToolError struct {
	Message string
}

func (e *ToolError) Error() string {
	return e.Message
}

// NewToolError 创建工具错误
func NewToolError(message string) *ToolError {
	return &ToolError{Message: message}
}

// TaskFieldUpdater 更新任务字段
type TaskFieldUpdater func(task *task.Task, fields map[string]interface{}) error

// StepFieldUpdater 更新步骤字段
type StepFieldUpdater func(step *task.TaskStep, fields map[string]interface{}) error

// ApplyTaskFields 应用任务字段更新
func ApplyTaskFields(task *task.Task, fields map[string]interface{}) error {
	for key, value := range fields {
		switch key {
		case "title":
			if title, ok := value.(string); ok {
				task.Title = title
			}
		case "description":
			if desc, ok := value.(string); ok {
				task.Description = desc
			}
		case "status":
			if status, ok := value.(string); ok {
				task.Status = status
			}
		case "priority":
			if priority, ok := value.(string); ok {
				task.Priority = priority
			}
		case "due_at":
			if dueAt, ok := value.(string); ok && dueAt != "" {
				parsedTime, err := time.Parse(time.RFC3339, dueAt)
				if err != nil {
					return err
				}
				task.DueAt = &parsedTime
			}
		}
	}
	return nil
}

// ApplyStepFields 应用步骤字段更新
func ApplyStepFields(step *task.TaskStep, fields map[string]interface{}) error {
	for key, value := range fields {
		switch key {
		case "title":
			if title, ok := value.(string); ok {
				step.Title = title
			}
		case "detail":
			if detail, ok := value.(string); ok {
				step.Detail = detail
			}
		case "status":
			if status, ok := value.(string); ok {
				step.Status = status
			}
		case "blocking_reason":
			if reason, ok := value.(string); ok {
				step.BlockingReason = reason
			}
		case "estimate_minutes":
			if estimate, ok := value.(float64); ok {
				intEstimate := int(estimate)
				step.EstimateMin = &intEstimate
			}
		case "order_index":
			if order, ok := value.(float64); ok {
				step.OrderIndex = int(order)
			}
		case "planned_start":
			if plannedStart, ok := value.(string); ok && plannedStart != "" {
				parsedTime, err := time.Parse(time.RFC3339, plannedStart)
				if err != nil {
					return err
				}
				step.PlannedStart = &parsedTime
			}
		case "planned_end":
			if plannedEnd, ok := value.(string); ok && plannedEnd != "" {
				parsedTime, err := time.Parse(time.RFC3339, plannedEnd)
				if err != nil {
					return err
				}
				step.PlannedEnd = &parsedTime
			}
		}
	}
	return nil
}
