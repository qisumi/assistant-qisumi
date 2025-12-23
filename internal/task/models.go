package task

import "time"

type Task struct {
	ID           uint64     `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint64     `gorm:"column:user_id;not null" json:"userId"`
	Title        string     `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description  string     `gorm:"column:description;type:text" json:"description"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:'todo'" json:"status"`
	Priority     string     `gorm:"column:priority;type:varchar(20);default:'medium'" json:"priority"`
	IsFocusToday bool       `gorm:"column:is_focus_today;default:false" json:"isFocusToday"`
	DueAt        *time.Time `gorm:"column:due_at" json:"dueAt,omitempty"`
	CreatedFrom  string     `gorm:"column:created_from;type:text" json:"createdFrom,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	Steps []TaskStep `gorm:"foreignKey:TaskID" json:"steps,omitempty"`
}

func (Task) TableName() string { return "tasks" }

type TaskStep struct {
	ID             uint64     `gorm:"primaryKey;column:id" json:"id"`
	TaskID         uint64     `gorm:"column:task_id;not null" json:"taskId"`
	OrderIndex     int        `gorm:"column:order_index;not null;default:0" json:"orderIndex"`
	Title          string     `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Detail         string     `gorm:"column:detail;type:text" json:"detail"`
	Status         string     `gorm:"column:status;type:varchar(20);not null;default:'todo'" json:"status"`
	BlockingReason string     `gorm:"column:blocking_reason;type:text" json:"blockingReason"`
	EstimateMin    *int       `gorm:"column:estimate_minutes" json:"estimateMinutes,omitempty"`
	PlannedStart   *time.Time `gorm:"column:planned_start" json:"plannedStart,omitempty"`
	PlannedEnd     *time.Time `gorm:"column:planned_end" json:"plannedEnd,omitempty"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (TaskStep) TableName() string { return "task_steps" }

type TaskDependency struct {
	ID uint64 `gorm:"primaryKey;column:id" json:"id"`

	PredecessorTaskID uint64  `gorm:"column:predecessor_task_id;not null" json:"predecessorTaskId"`
	PredecessorStepID *uint64 `gorm:"column:predecessor_step_id" json:"predecessorStepId,omitempty"`
	SuccessorTaskID   uint64  `gorm:"column:successor_task_id;not null" json:"successorTaskId"`
	SuccessorStepID   *uint64 `gorm:"column:successor_step_id" json:"successorStepId,omitempty"`

	Condition string    `gorm:"column:dependency_condition;type:varchar(20);not null" json:"condition"`
	Action    string    `gorm:"column:action;type:varchar(20);not null;default:'unlock_step'" json:"action"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (TaskDependency) TableName() string { return "task_dependencies" }

// --- Patch/Update related structs ---

type UpdateTaskFields struct {
	Title        *string `json:"title,omitempty"`
	Description  *string `json:"description,omitempty"`
	Status       *string `json:"status,omitempty"`   // "todo" | "in_progress" | "done" | "cancelled"
	Priority     *string `json:"priority,omitempty"` // "low" | "medium" | "high"
	IsFocusToday *bool   `json:"isFocusToday,omitempty"`
	DueAt        *string `json:"dueAt,omitempty"` // ISO 8601
}

type UpdateStepFields struct {
	Title          *string `json:"title,omitempty"`
	Detail         *string `json:"detail,omitempty"`
	Status         *string `json:"status,omitempty"` // "locked" | "todo" | "in_progress" | "done" | "blocked"
	BlockingReason *string `json:"blockingReason,omitempty"`
	EstimateMin    *int    `json:"estimateMinutes,omitempty"`
	OrderIndex     *int    `json:"orderIndex,omitempty"`
	PlannedStart   *string `json:"plannedStart,omitempty"` // ISO 8601
	PlannedEnd     *string `json:"plannedEnd,omitempty"`   // ISO 8601
}

type NewStepRecord struct {
	Title             string  `json:"title"`
	Detail            string  `json:"detail"`
	EstimateMinutes   *int    `json:"estimateMinutes,omitempty"`
	InsertAfterStepID *uint64 `json:"insertAfterStepId,omitempty"`
}

type DependencyItem struct {
	PredecessorTaskID uint64  `json:"predecessorTaskId"`
	PredecessorStepID *uint64 `json:"predecessorStepId,omitempty"`
	SuccessorTaskID   uint64  `json:"successorTaskId"`
	SuccessorStepID   *uint64 `json:"successorStepId,omitempty"`
	Condition         string  `json:"condition"` // "task_done" | "step_done"
	Action            string  `json:"action"`    // "unlock_step" | "set_task_todo" | "notify_only"
}
