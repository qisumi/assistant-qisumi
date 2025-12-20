package db

import (
	"gorm.io/gorm"

	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"
)

// AutoMigrate 执行 GORM 自动迁移
// 一般在 main 启动时调用一次即可。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&auth.User{},
		&auth.UserLLMSetting{},
		&task.Task{},
		&task.TaskStep{},
		&task.TaskDependency{},
		&session.Session{},
		&session.Message{},
	)
}
