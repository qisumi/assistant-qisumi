package task

import "time"

type Task struct {
	ID          uint64     `json:"id"`
	UserID      uint64     `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Steps       []Step     `json:"steps,omitempty"`
}

type Step struct {
	ID             uint64     `json:"id"`
	TaskID         uint64     `json:"task_id"`
	OrderIndex     int        `json:"order_index"`
	Title          string     `json:"title"`
	Detail         string     `json:"detail"`
	Status         string     `json:"status"`
	BlockingReason string     `json:"blocking_reason"`
	EstimateMin    *int       `json:"estimate_minutes,omitempty"`
	PlannedStart   *time.Time `json:"planned_start,omitempty"`
	PlannedEnd     *time.Time `json:"planned_end,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}