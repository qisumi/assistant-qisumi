package http

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"assistant-qisumi/internal/config"
)

// Server HTTP服务器
type Server struct {
	engine *gin.Engine
	db     *sql.DB
	cfg    config.HTTPConfig
}

// NewServer 创建新的HTTP服务器
func NewServer(cfg config.HTTPConfig, db *sql.DB) *Server {
	engine := gin.Default()

	s := &Server{
		engine: engine,
		db:     db,
		cfg:    cfg,
	}

	// 注册路由
	s.setupRoutes()

	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// API路由组
	api := s.engine.Group("/api")
	{
		// 健康检查
		api.GET("/health", s.healthCheck)
	}
}

// healthCheck 健康检查处理器
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	return s.engine.Run(addr)
}
