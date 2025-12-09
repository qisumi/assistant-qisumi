package test

import (
	"testing"

	"assistant-qisumi/internal/db"
)

// TestDBConnection 测试数据库连接
func TestDBConnection(t *testing.T) {
	cfg, err := GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// 初始化数据库连接
	dbConn, err := db.InitDB(cfg.DB)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// 测试数据库连接
	if err := dbConn.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	t.Log("Database connection successful")
}
