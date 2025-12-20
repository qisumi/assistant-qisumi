package db

import (
	"database/sql"
	"fmt"

	"assistant-qisumi/internal/config"

	_ "github.com/glebarez/go-sqlite"
	"github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB 初始化原生数据库连接
func InitDB(cfg config.DBConfig) (*sql.DB, error) {
	var driverName, dsn string
	if cfg.Type == "sqlite" {
		driverName = "sqlite"
		dsn = cfg.FilePath
	} else {
		driverName = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	}
	return sql.Open(driverName, dsn)
}

// InitGORM 从现有的 sql.DB 初始化 GORM
func InitGORM(sqlDB *sql.DB, dbType string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if dbType == "sqlite" {
		dialector = &sqlite.Dialector{
			Conn: sqlDB,
		}
	} else {
		dialector = mysql.New(mysql.Config{
			Conn: sqlDB,
		})
	}
	return gorm.Open(dialector, &gorm.Config{})
}
