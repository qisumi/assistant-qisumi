package auth

import "time"

type User struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	Email        string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`
	DisplayName  string    `gorm:"column:display_name;type:varchar(255)" json:"display_name"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (User) TableName() string { return "users" }

type UserLLMSetting struct {
	ID               uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID           uint64    `gorm:"column:user_id;not null;index" json:"user_id"`
	BaseURL          string    `gorm:"column:base_url;type:varchar(512);not null" json:"base_url"`
	APIKeyEnc        string    `gorm:"column:api_key_enc;type:text;not null" json:"api_key_enc"`
	Model            string    `gorm:"column:model;type:varchar(255);not null" json:"model"`
	ThinkingType     string    `gorm:"column:thinking_type;type:varchar(20);default:'auto'" json:"thinking_type"`
	ReasoningEffort  string    `gorm:"column:reasoning_effort;type:varchar(20);default:'medium'" json:"reasoning_effort"`
	EnableThinking   bool      `gorm:"column:enable_thinking;default:false" json:"enable_thinking"`
	AssistantName    string    `gorm:"column:assistant_name;type:varchar(100);default:'小奇'" json:"assistant_name"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserLLMSetting) TableName() string { return "user_llm_settings" }
