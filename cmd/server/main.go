package main

import (
	"log"

	"assistant-qisumi/internal/config"
	"assistant-qisumi/internal/db"
	"assistant-qisumi/internal/http"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	db, err := db.InitDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 初始化HTTP服务器
	server := http.NewServer(cfg.HTTP, cfg.JWT, cfg.Crypto, db)

	// 启动服务器
	log.Printf("Server starting on %s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}