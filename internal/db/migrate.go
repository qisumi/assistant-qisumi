package db

import (
	"gorm.io/gorm"

	"assistant-qisumi/internal/domain"
)

// AutoMigrate 执行 GORM 自动迁移
// 一般在 main 启动时调用一次即可。
// 现在直接使用 domain 包的模型，避免循环依赖
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.UserLLMSetting{},
		&domain.Task{},
		&domain.TaskStep{},
		&domain.TaskDependency{},
		&domain.Session{},
		&domain.Message{},
	)
}
