package http

import (
	"net/http"

	"assistant-qisumi/internal/auth"

	"github.com/gin-gonic/gin"
)

// SettingsHandler 处理用户设置相关请求
type SettingsHandler struct {
	llmSettingService *auth.LLMSettingService
}

// NewSettingsHandler 创建新的设置处理器
func NewSettingsHandler(llmSettingService *auth.LLMSettingService) *SettingsHandler {
	return &SettingsHandler{
		llmSettingService: llmSettingService,
	}
}

// RegisterRoutes 注册设置相关路由
func (h *SettingsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/settings/llm", h.getLLMSettings)
	rg.POST("/settings/llm", h.updateLLMSettings)
	rg.DELETE("/settings/llm", h.deleteLLMSettings)
}

// getLLMSettings 获取当前用户的LLM设置
func (h *SettingsHandler) getLLMSettings(c *gin.Context) {
	userID := GetUserID(c)

	config, err := h.llmSettingService.GetLLMConfig(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if config == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, config)
}

// updateLLMSettings 更新当前用户的LLM设置
func (h *SettingsHandler) updateLLMSettings(c *gin.Context) {
	userID := GetUserID(c)

	var req auth.LLMSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.llmSettingService.UpdateLLMSetting(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取更新后的设置返回
	config, _ := h.llmSettingService.GetLLMConfig(c.Request.Context(), userID)
	c.JSON(http.StatusOK, config)
}

// deleteLLMSettings 删除当前用户的LLM设置
func (h *SettingsHandler) deleteLLMSettings(c *gin.Context) {
	userID := GetUserID(c)

	if err := h.llmSettingService.DeleteLLMSetting(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "LLM settings deleted",
	})
}
