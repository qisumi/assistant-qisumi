package db

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormDB 初始化 GORM，控制连接池，适配 2C / 2G 小机
func NewGormDB(dbType string, dsn string) (*gorm.DB, error) {
	cfg := &gorm.Config{
		// 日志级别你可以按需调整
		Logger: logger.Default.LogMode(logger.Warn),
		// 可以开启 PreparedStmt 优化查询
		PrepareStmt: true,
	}

	var dialector gorm.Dialector
	if dbType == "sqlite" {
		dialector = sqlite.Open(dsn)
	} else {
		dialector = mysql.Open(dsn)
	}

	db, err := gorm.Open(dialector, cfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 小机器：连接池尽量保守
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
