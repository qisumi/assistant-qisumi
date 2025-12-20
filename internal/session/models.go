package session

import "time"

type Session struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;index" json:"user_id"`
	TaskID    *uint64   `gorm:"column:task_id;index" json:"task_id,omitempty"`
	Type      string    `gorm:"column:type;type:varchar(20);not null;default:'task'" json:"type"` // "task" or "global"
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Session) TableName() string { return "sessions" }

type Message struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`
	SessionID uint64    `gorm:"column:session_id;not null;index" json:"session_id"`
	Role      string    `gorm:"column:role;type:varchar(20);not null" json:"role"` // "user" | "assistant" | "system"
	AgentName *string   `gorm:"column:agent_name;type:varchar(64)" json:"agent_name,omitempty"`
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Message) TableName() string { return "messages" }
