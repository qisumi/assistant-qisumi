package main

import (
	"fmt"
	"log"

	"assistant-qisumi/internal/config"
	"assistant-qisumi/internal/db"
	"assistant-qisumi/internal/http"
	"assistant-qisumi/internal/logger"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 获取可执行文件所在目录用于日志
	execDir, err := config.GetExecutableDir()
	if err != nil {
		log.Fatalf("Failed to get executable directory: %v", err)
	}

	// 初始化logger
	if err := logger.Init(execDir, cfg.Log.Level); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Logger.Sync()

	// 初始化数据库连接
	var dsn string
	if cfg.DB.Type == "sqlite" {
		dsn = cfg.DB.FilePath
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)
	}

	gormDB, err := db.NewGormDB(cfg.DB.Type, dsn)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// 执行自动迁移
	if err := db.AutoMigrate(gormDB); err != nil {
		logger.Logger.Fatal("Failed to auto migrate", zap.Error(err))
	}

	// 获取 sql.DB 用于 defer close (虽然 GORM 会管理，但保持习惯)
	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Logger.Fatal("Failed to get sql.DB", zap.Error(err))
	}
	defer sqlDB.Close()

	// 初始化HTTP服务器
	server := http.NewServer(cfg.HTTP, cfg.JWT, cfg.Crypto, cfg.LLM, gormDB, nil)

	// 启动服务器
	logger.Logger.Info("Server starting",
		zap.String("host", cfg.HTTP.Host),
		zap.String("port", cfg.HTTP.Port),
	)
	if err := server.Start(); err != nil {
		logger.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}
