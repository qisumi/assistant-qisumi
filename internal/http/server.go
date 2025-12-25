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
	llmCfg    config.LLMConfig
	llmClient llm.Client
}

// NewServer 创建新的HTTP服务器
func NewServer(cfg config.HTTPConfig, jwtCfg config.JWTConfig, cryptoCfg config.CryptoConfig, llmCfg config.LLMConfig, db *gorm.DB, llmClient llm.Client) *Server {
	engine := gin.Default()

	if llmClient == nil {
		llmClient = llm.NewHTTPClient()
	}

	s := &Server{
		engine:    engine,
		db:        db,
		cfg:       cfg,
		jwtCfg:    jwtCfg,
		cryptoCfg: cryptoCfg,
		llmCfg:    llmCfg,
		llmClient: llmClient,
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
		defaultLLMConfig := &auth.LLMConfig{
			BaseURL:         s.llmCfg.APIBaseURL,
			APIKey:          s.llmCfg.APIKey,
			Model:           s.llmCfg.ModelName,
			ThinkingType:    s.llmCfg.ThinkingType,
			ReasoningEffort: s.llmCfg.ReasoningEffort,
			HasAPIKey:       s.llmCfg.APIKey != "",
			IsDefault:       true,
		}
		llmSettingService := auth.NewLLMSettingService(llmSettingRepo, s.cryptoCfg.APIKeyEncryptionKey, defaultLLMConfig)

		taskRepo := task.NewRepository(s.db)
		taskSvc := task.NewService(taskRepo, s.llmClient)

		sessionRepo := session.NewRepository(s.db)

		// Dependency Service
		dependencySvc := dependency.NewService(s.db, taskRepo, sessionRepo)

		// Agents
		router := agent.NewSimpleRouter()

		// 初始化工具执行器映射（提前创建供所有agent使用）
		toolMap := agent.NewToolExecutors()

		// 初始化Chat Completions处理器（提前创建供所有agent使用）
		chatCompletionsHandler := agent.NewChatCompletionsHandler(s.llmClient, toolMap)

		// 创建agents，传入chatCompletionsHandler
		executorAgent := agent.NewExecutorAgent(s.llmClient, chatCompletionsHandler)
		plannerAgent := agent.NewPlannerAgent(s.llmClient, chatCompletionsHandler)
		summarizerAgent := agent.NewSummarizerAgent(s.llmClient)
		globalAgent := agent.NewGlobalAgent(s.llmClient, chatCompletionsHandler)
		taskCreationAgent := agent.NewTaskCreationAgent(s.llmClient)
		agents := []agent.Agent{executorAgent, plannerAgent, summarizerAgent, globalAgent, taskCreationAgent}
		agentSvc := agent.NewService(router, agents, taskRepo, sessionRepo, dependencySvc, s.db, s.llmClient)

		// 初始化处理器
		authHandler := NewAuthHandler(authSvc)
		taskHandler := NewTaskHandler(taskSvc, sessionRepo, llmSettingService)
		sessionHandler := NewSessionHandler(agentSvc, sessionRepo, llmSettingService)
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

// ServeHTTP 实现 http.Handler 接口，方便测试
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}
