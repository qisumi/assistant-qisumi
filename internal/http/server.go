package http

import (
	"fmt"
	"net/http"

	"assistant-qisumi/internal/agent"
	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/config"
	"assistant-qisumi/internal/dependency"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Server HTTP服务器
type Server struct {
	engine    *gin.Engine
	db        *gorm.DB
	cfg       config.HTTPConfig
	jwtCfg    config.JWTConfig
	cryptoCfg config.CryptoConfig
}

// NewServer 创建新的HTTP服务器
func NewServer(cfg config.HTTPConfig, jwtCfg config.JWTConfig, cryptoCfg config.CryptoConfig, db *gorm.DB) *Server {
	engine := gin.Default()

	s := &Server{
		engine:    engine,
		db:        db,
		cfg:       cfg,
		jwtCfg:    jwtCfg,
		cryptoCfg: cryptoCfg,
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

		// 初始化服务
		jwtMgr := auth.NewJWTManager(s.jwtCfg.Secret)
		authSvc := auth.NewService(s.db, jwtMgr)

		// LLM 设置服务
		llmSettingRepo := auth.NewLLMSettingRepository(s.db)
		llmSettingService := auth.NewLLMSettingService(llmSettingRepo, s.cryptoCfg.APIKeyEncryptionKey)

		taskRepo := task.NewRepository(s.db)
		llmClient := llm.NewHTTPClient()
		taskSvc := task.NewService(taskRepo, llmClient)

		sessionRepo := session.NewRepository(s.db)

		// Dependency Service
		dependencySvc := dependency.NewService(s.db, taskRepo, sessionRepo)

		// Agents
		router := agent.NewSimpleRouter()
		executorAgent := agent.NewExecutorAgent(llmClient)
		plannerAgent := agent.NewPlannerAgent(llmClient)
		summarizerAgent := agent.NewSummarizerAgent(llmClient)
		globalAgent := agent.NewGlobalAgent(llmClient)
		taskCreationAgent := agent.NewTaskCreationAgent(llmClient)
		agents := []agent.Agent{executorAgent, plannerAgent, summarizerAgent, globalAgent, taskCreationAgent}
		agentSvc := agent.NewService(router, agents, taskRepo, sessionRepo, dependencySvc, s.db, llmClient)

		// 初始化处理器
		authHandler := NewAuthHandler(authSvc)
		taskHandler := NewTaskHandler(taskSvc, llmSettingService)
		sessionHandler := NewSessionHandler(agentSvc, llmSettingService)
		settingsHandler := NewSettingsHandler(llmSettingService)

		// 认证路由
		authHandler.RegisterRoutes(api.Group("/auth"))

		// 需要登录的路由
		authGroup := api.Group("")
		authGroup.Use(AuthMiddleware(jwtMgr))

		// 任务路由
		taskHandler.RegisterRoutes(authGroup)

		// 会话路由
		sessionHandler.RegisterRoutes(authGroup)

		// 设置路由
		settingsHandler.RegisterRoutes(authGroup)
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
