package session

import "time"

type Session struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	TaskID    *uint64   `json:"task_id,omitempty"`
	Type      string    `json:"type"` // "task" or "global"
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        uint64    `json:"id"`
	SessionID uint64    `json:"session_id"`
	Role      string    `json:"role"` // "user" | "assistant" | "system"
	AgentName *string   `json:"agent_name,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}