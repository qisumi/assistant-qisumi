package db

import (
	"database/sql"
	"fmt"

	"assistant-qisumi/internal/config"

	_ "github.com/go-sql-driver/mysql" // 确保已执行 go get github.com/go-sql-driver/mysql
)

// InitDB 初始化数据库连接
func InitDB(cfg config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	return db, nil
}
