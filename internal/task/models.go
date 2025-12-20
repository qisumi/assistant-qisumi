package task

import "time"

type Task struct {
	ID          uint64     `gorm:"primaryKey;column:id" json:"id"`
	UserID      uint64     `gorm:"column:user_id;not null" json:"user_id"`
	Title       string     `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	Status      string     `gorm:"column:status;type:enum('todo','in_progress','done','cancelled');not null;default:'todo'" json:"status"`
	Priority    string     `gorm:"column:priority;type:enum('low','medium','high');default:'medium'" json:"priority"`
	DueAt       *time.Time `gorm:"column:due_at" json:"due_at,omitempty"`
	CreatedFrom string     `gorm:"column:created_from;type:text" json:"created_from,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Steps []TaskStep `gorm:"foreignKey:TaskID" json:"steps,omitempty"`
}

func (Task) TableName() string { return "tasks" }

type TaskStep struct {
	ID             uint64     `gorm:"primaryKey;column:id" json:"id"`
	TaskID         uint64     `gorm:"column:task_id;not null" json:"task_id"`
	OrderIndex     int        `gorm:"column:order_index;not null;default:0" json:"order_index"`
	Title          string     `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Detail         string     `gorm:"column:detail;type:text" json:"detail"`
	Status         string     `gorm:"column:status;type:enum('locked','todo','in_progress','done','blocked');not null;default:'todo'" json:"status"`
	BlockingReason string     `gorm:"column:blocking_reason;type:text" json:"blocking_reason"`
	EstimateMin    *int       `gorm:"column:estimate_minutes" json:"estimate_minutes,omitempty"`
	PlannedStart   *time.Time `gorm:"column:planned_start" json:"planned_start,omitempty"`
	PlannedEnd     *time.Time `gorm:"column:planned_end" json:"planned_end,omitempty"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (TaskStep) TableName() string { return "task_steps" }

type TaskDependency struct {
	ID uint64 `gorm:"primaryKey;column:id" json:"id"`

	PredecessorTaskID uint64  `gorm:"column:predecessor_task_id;not null" json:"predecessor_task_id"`
	PredecessorStepID *uint64 `gorm:"column:predecessor_step_id" json:"predecessor_step_id,omitempty"`
	SuccessorTaskID   uint64  `gorm:"column:successor_task_id;not null" json:"successor_task_id"`
	SuccessorStepID   *uint64 `gorm:"column:successor_step_id" json:"successor_step_id,omitempty"`

	Condition string    `gorm:"column:dependency_condition;type:enum('task_done','step_done');not null" json:"condition"`
	Action    string    `gorm:"column:action;type:enum('unlock_step','set_task_todo','notify_only');not null;default:'unlock_step'" json:"action"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (TaskDependency) TableName() string { return "task_dependencies" }

// --- Patch/Update related structs ---

type UpdateTaskFields struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`   // "todo" | "in_progress" | "done" | "cancelled"
	Priority    *string `json:"priority,omitempty"` // "low" | "medium" | "high"
	DueAt       *string `json:"due_at,omitempty"`   // ISO 8601
}

type UpdateStepFields struct {
	Title          *string `json:"title,omitempty"`
	Detail         *string `json:"detail,omitempty"`
	Status         *string `json:"status,omitempty"` // "locked" | "todo" | "in_progress" | "done" | "blocked"
	BlockingReason *string `json:"blocking_reason,omitempty"`
	EstimateMin    *int    `json:"estimate_minutes,omitempty"`
	OrderIndex     *int    `json:"order_index,omitempty"`
	PlannedStart   *string `json:"planned_start,omitempty"` // ISO 8601
	PlannedEnd     *string `json:"planned_end,omitempty"`   // ISO 8601
}

type NewStepRecord struct {
	Title             string  `json:"title"`
	Detail            string  `json:"detail"`
	EstimateMinutes   *int    `json:"estimate_minutes,omitempty"`
	InsertAfterStepID *uint64 `json:"insert_after_step_id,omitempty"`
}

type DependencyItem struct {
	PredecessorTaskID uint64  `json:"predecessor_task_id"`
	PredecessorStepID *uint64 `json:"predecessor_step_id,omitempty"`
	SuccessorTaskID   uint64  `json:"successor_task_id"`
	SuccessorStepID   *uint64 `json:"successor_step_id,omitempty"`
	Condition         string  `json:"condition"` // "task_done" | "step_done"
	Action            string  `json:"action"`    // "unlock_step" | "set_task_todo" | "notify_only"
}
