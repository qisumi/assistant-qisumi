package http

import (
	"net/http"
	"strconv"

	"assistant-qisumi/internal/agent"
	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/llm"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	agentSvc      *agent.Service
	llmSettingSvc *auth.LLMSettingService
}

func NewSessionHandler(agentSvc *agent.Service, llmSettingSvc *auth.LLMSettingService) *SessionHandler {
	return &SessionHandler{
		agentSvc:      agentSvc,
		llmSettingSvc: llmSettingSvc,
	}
}

func (h *SessionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/sessions/:id/messages", h.postMessage)
}

type postMessageReq struct {
	Content string `json:"content" binding:"required"`
}

func (h *SessionHandler) postMessage(c *gin.Context) {
	userID := GetUserID(c)
	sidStr := c.Param("id")
	sid, err := strconv.ParseUint(sidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var req postMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从 DB 根据 userID 获取 LLMConfig
	llmConfig, err := h.llmSettingSvc.GetLLMConfig(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get LLM config"})
		return
	}
	if llmConfig == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LLM config not set"})
		return
	}

	// 转换为 llm.Config 类型
	cfg := llm.Config{
		BaseURL: llmConfig.BaseURL,
		APIKey:  llmConfig.APIKey,
		Model:   llmConfig.Model,
	}

	resp, err := h.agentSvc.HandleUserMessage(c, userID, sid, req.Content, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"assistant_message": resp.AssistantMessage,
		"task_patches":      resp.TaskPatches,
	})
}
