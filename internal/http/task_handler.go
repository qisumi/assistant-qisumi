package http

import (
	"net/http"

	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/task"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskSvc       *task.Service
	llmSettingSvc *auth.LLMSettingService
}

func NewTaskHandler(taskSvc *task.Service, llmSettingSvc *auth.LLMSettingService) *TaskHandler {
	return &TaskHandler{
		taskSvc:       taskSvc,
		llmSettingSvc: llmSettingSvc,
	}
}

func (h *TaskHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tasks/from-text", h.createFromText)
	// TODO: 添加 tasks 列表 / 详情 / patch 路由
}

type createFromTextReq struct {
	RawText string `json:"raw_text" binding:"required"`
}

func (h *TaskHandler) createFromText(c *gin.Context) {
	userID := GetUserID(c)
	var req createFromTextReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从 DB 根据 userID 查 LLMConfig
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

	t, err := h.taskSvc.CreateFromText(c, userID, req.RawText, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": t})
}

// listTasks 获取任务列表
func (h *TaskHandler) listTasks(c *gin.Context) {
	userID := GetUserID(c)
	// TODO: 实现任务列表查询
	tasks, err := h.taskSvc.ListTasks(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// getTask 获取任务详情
func (h *TaskHandler) getTask(c *gin.Context) {
	userID := GetUserID(c)
	idStr := c.Param("id")
	// TODO: 解析ID并实现任务详情查询
	// 暂时返回占位响应
	c.JSON(http.StatusOK, gin.H{"message": "getTask handler", "user_id": userID, "task_id": idStr})
}

// patchTask 更新任务
func (h *TaskHandler) patchTask(c *gin.Context) {
	userID := GetUserID(c)
	idStr := c.Param("id")
	// TODO: 解析ID并实现任务更新
	// 暂时返回占位响应
	c.JSON(http.StatusOK, gin.H{"message": "patchTask handler", "user_id": userID, "task_id": idStr})
}

// createTask 创建任务
func (h *TaskHandler) createTask(c *gin.Context) {
	userID := GetUserID(c)
	// TODO: 实现任务创建
	// 暂时返回占位响应
	c.JSON(http.StatusOK, gin.H{"message": "createTask handler", "user_id": userID})
}
