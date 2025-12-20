package http

import (
	"net/http"
	"strconv"

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
	rg.GET("/tasks", h.listTasks)
	rg.POST("/tasks", h.createTask)
	rg.GET("/tasks/:id", h.getTask)
	rg.PATCH("/tasks/:id", h.patchTask)
}

type CreateFromTextReq struct {
	RawText string `json:"raw_text" binding:"required"`
}

func (h *TaskHandler) createFromText(c *gin.Context) {
	userID := GetUserID(c)
	var req CreateFromTextReq
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
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	t, err := h.taskSvc.GetTask(c, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": t})
}

// patchTask 更新任务
func (h *TaskHandler) patchTask(c *gin.Context) {
	userID := GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var fields task.UpdateTaskFields
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskSvc.UpdateTask(c, userID, id, fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task updated"})
}

// createTask 创建任务
func (h *TaskHandler) createTask(c *gin.Context) {
	userID := GetUserID(c)
	var t task.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t.UserID = userID

	if err := h.taskSvc.CreateTask(c, &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": t})
}
