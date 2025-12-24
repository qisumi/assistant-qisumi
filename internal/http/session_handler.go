package http

import (
	"errors"

	"assistant-qisumi/internal/agent"
	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/session"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// validateSessionOwner 验证 session 属于该用户
func (h *SessionHandler) validateSessionOwner(c *gin.Context, sid, userID uint64) error {
	sess, err := h.sessionRepo.GetSession(c, sid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			R.NotFound(c, "session not found")
			return err
		}
		R.InternalError(c, err.Error())
		return err
	}
	if sess.UserID != userID {
		R.Forbidden(c, "forbidden")
		return errors.New("forbidden")
	}
	return nil
}

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
		R.InternalError(c, err.Error())
		return
	}
	R.Success(c, gin.H{"session": sess})
}

func (h *SessionHandler) listMessages(c *gin.Context) {
	userID := GetUserID(c)
	sid, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	// 验证 session 属于该用户
	if err := h.validateSessionOwner(c, sid, userID); err != nil {
		return
	}

	messages, err := h.sessionRepo.ListRecentMessages(c, sid, 50)
	if err != nil {
		R.InternalError(c, err.Error())
		return
	}

	R.Success(c, gin.H{
		"sessionId": sid,
		"messages":  messages,
	})
}

type PostMessageReq struct {
	Content string `json:"content" binding:"required"`
}

func (h *SessionHandler) postMessage(c *gin.Context) {
	userID := GetUserID(c)
	sid, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	// 验证 session 存在且属于该用户
	if err := h.validateSessionOwner(c, sid, userID); err != nil {
		return
	}

	var req PostMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		R.BadRequest(c, err.Error())
		return
	}

	cfg, err := GetLLMConfig(c, h.llmSettingSvc, userID)
	if err != nil {
		return
	}

	resp, err := h.agentSvc.HandleUserMessage(c, userID, sid, req.Content, *cfg)
	if err != nil {
		R.InternalError(c, "HandleUserMessage failed: "+err.Error())
		return
	}

	R.Success(c, gin.H{
		"sessionId":        sid,
		"assistantMessage": resp.AssistantMessage,
		"taskPatches":      resp.TaskPatches,
	})
}

func (h *SessionHandler) clearMessages(c *gin.Context) {
	userID := GetUserID(c)
	sid, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	// 验证 session 属于该用户
	if err := h.validateSessionOwner(c, sid, userID); err != nil {
		return
	}

	// 清空该 session 的所有消息
	if err := h.sessionRepo.ClearMessages(c, sid); err != nil {
		R.InternalError(c, err.Error())
		return
	}

	R.Success(c, gin.H{"success": true})
}
