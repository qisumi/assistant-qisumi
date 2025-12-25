// Package domain 定义核心领域模型
// 这个包作为基础层，不依赖任何其他 internal 包，避免循环依赖
package domain

import "time"

// ==================== User 相关模型 ====================

type User struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	Email        string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`
	DisplayName  string    `gorm:"column:display_name;type:varchar(255)" json:"display_name"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (User) TableName() string { return "users" }

type UserLLMSetting struct {
	ID              uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID          uint64    `gorm:"column:user_id;not null;index" json:"user_id"`
	BaseURL         string    `gorm:"column:base_url;type:varchar(512);not null" json:"base_url"`
	APIKeyEnc       string    `gorm:"column:api_key_enc;type:text;not null" json:"api_key_enc"`
	Model           string    `gorm:"column:model;type:varchar(255);not null" json:"model"`
	ThinkingType    string    `gorm:"column:thinking_type;type:varchar(20);default:'auto'" json:"thinking_type"`
	ReasoningEffort string    `gorm:"column:reasoning_effort;type:varchar(20);default:'medium'" json:"reasoning_effort"`
	EnableThinking  bool      `gorm:"column:enable_thinking;default:false" json:"enable_thinking"`
	AssistantName   string    `gorm:"column:assistant_name;type:varchar(100);default:'小奇'" json:"assistant_name"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserLLMSetting) TableName() string { return "user_llm_settings" }

// ==================== Session 相关模型 ====================

type Session struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;index" json:"userId"`
	TaskID    *uint64   `gorm:"column:task_id;index" json:"taskId,omitempty"`
	Type      string    `gorm:"column:type;type:varchar(20);not null;default:'task'" json:"type"` // "task" or "global"
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Session) TableName() string { return "sessions" }

type Message struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`
	SessionID uint64    `gorm:"column:session_id;not null;index" json:"sessionId"`
	Role      string    `gorm:"column:role;type:varchar(20);not null" json:"role"` // "user" | "assistant" | "system"
	AgentName *string   `gorm:"column:agent_name;type:varchar(64)" json:"agentName,omitempty"`
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Message) TableName() string { return "messages" }

// ==================== Task 相关模型 ====================

type Task struct {
	ID           uint64        `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint64        `gorm:"column:user_id;not null" json:"userId"`
	Title        string        `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description  string        `gorm:"column:description;type:text" json:"description"`
	Status       string        `gorm:"column:status;type:varchar(20);not null;default:'todo'" json:"status"`
	Priority     string        `gorm:"column:priority;type:varchar(20);default:'medium'" json:"priority"`
	IsFocusToday bool          `gorm:"column:is_focus_today;default:false" json:"isFocusToday"`
	DueAt        *FlexibleTime `gorm:"column:due_at" json:"dueAt,omitempty"`
	CreatedFrom  string        `gorm:"column:created_from;type:text" json:"createdFrom,omitempty"`
	CreatedAt    time.Time     `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	CompletedAt  *time.Time    `gorm:"column:completed_at" json:"completedAt,omitempty"`

	Steps []TaskStep `gorm:"foreignKey:TaskID" json:"steps,omitempty"`
}

func (Task) TableName() string { return "tasks" }

type TaskStep struct {
	ID             uint64        `gorm:"primaryKey;column:id" json:"id"`
	TaskID         uint64        `gorm:"column:task_id;not null" json:"taskId"`
	OrderIndex     int           `gorm:"column:order_index;not null;default:0" json:"orderIndex"`
	Title          string        `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Detail         string        `gorm:"column:detail;type:text" json:"detail"`
	Status         string        `gorm:"column:status;type:varchar(20);not null;default:'todo'" json:"status"`
	BlockingReason string        `gorm:"column:blocking_reason;type:text" json:"blockingReason"`
	EstimateMin    *int          `gorm:"column:estimate_minutes" json:"estimateMinutes,omitempty"`
	PlannedStart   *FlexibleTime `gorm:"column:planned_start" json:"plannedStart,omitempty"`
	PlannedEnd     *FlexibleTime `gorm:"column:planned_end" json:"plannedEnd,omitempty"`
	CreatedAt      time.Time     `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	CompletedAt    *time.Time    `gorm:"column:completed_at" json:"completedAt,omitempty"`
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

// ==================== Task 更新相关结构 ====================

type UpdateTaskFields struct {
	Title        *string `json:"title,omitempty"`
	Description  *string `json:"description,omitempty"`
	Status       *string `json:"status,omitempty"`   // "todo" | "in_progress" | "done" | "cancelled"
	Priority     *string `json:"priority,omitempty"` // "low" | "medium" | "high"
	IsFocusToday *bool   `json:"isFocusToday,omitempty"`
	DueAt        *string `json:"dueAt,omitempty"`       // RFC3339
	CompletedAt  *string `json:"completedAt,omitempty"` // RFC3339
}

type UpdateStepFields struct {
	Title          *string `json:"title,omitempty"`
	Detail         *string `json:"detail,omitempty"`
	Status         *string `json:"status,omitempty"` // "locked" | "todo" | "in_progress" | "done" | "blocked"
	BlockingReason *string `json:"blockingReason,omitempty"`
	EstimateMin    *int    `json:"estimateMinutes,omitempty"`
	OrderIndex     *int    `json:"orderIndex,omitempty"`
	PlannedStart   *string `json:"plannedStart,omitempty"` // RFC3339
	PlannedEnd     *string `json:"plannedEnd,omitempty"`   // RFC3339
	CompletedAt    *string `json:"completedAt,omitempty"`  // RFC3339
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
