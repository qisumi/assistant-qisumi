package http

import (
	"errors"
	"net/http"
	"strconv"

	"assistant-qisumi/internal/agent"
	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SessionHandler struct {
	agentSvc      *agent.Service
	sessionRepo   *session.Repository
	llmSettingSvc *auth.LLMSettingService
}

func NewSessionHandler(agentSvc *agent.Service, sessionRepo *session.Repository, llmSettingSvc *auth.LLMSettingService) *SessionHandler {
	return &SessionHandler{
		agentSvc:      agentSvc,
		sessionRepo:   sessionRepo,
		llmSettingSvc: llmSettingSvc,
	}
}

func (h *SessionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/sessions/global", h.getGlobalSession)
	rg.GET("/sessions/:id/messages", h.listMessages)
	rg.POST("/sessions/:id/messages", h.postMessage)
	rg.DELETE("/sessions/:id/messages", h.clearMessages)
}

func (h *SessionHandler) getGlobalSession(c *gin.Context) {
	userID := GetUserID(c)
	sess, err := h.sessionRepo.GetGlobalSessionOrCreate(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"session": sess})
}

func (h *SessionHandler) listMessages(c *gin.Context) {
	userID := GetUserID(c)
	sidStr := c.Param("id")
	sid, err := strconv.ParseUint(sidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// 验证 session 属于该用户
	sess, err := h.sessionRepo.GetSession(c, sid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if sess.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	messages, err := h.sessionRepo.ListRecentMessages(c, sid, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sid,
		"messages":  messages,
	})
}

type PostMessageReq struct {
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

	sess, err := h.sessionRepo.GetSession(c, sid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sess.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req PostMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从 DB 根据 userID 获取 LLMConfig
	llmConfig, err := h.llmSettingSvc.GetLLMConfig(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get LLM config: " + err.Error()})
		return
	}
	if llmConfig == nil || llmConfig.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LLM API key not set. Please configure it in settings or contact administrator."})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "HandleUserMessage failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"sessionId":        sid,
		"assistantMessage": resp.AssistantMessage,
		"taskPatches":      resp.TaskPatches,
	})
}

func (h *SessionHandler) clearMessages(c *gin.Context) {
	userID := GetUserID(c)
	sidStr := c.Param("id")
	sid, err := strconv.ParseUint(sidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// 验证 session 属于该用户
	sess, err := h.sessionRepo.GetSession(c, sid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if sess.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// 清空该 session 的所有消息
	if err := h.sessionRepo.ClearMessages(c, sid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
