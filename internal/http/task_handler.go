package http

import (
	"errors"
	"net/http"
	"strconv"

	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct {
	taskSvc       *task.Service
	sessionRepo   *session.Repository
	llmSettingSvc *auth.LLMSettingService
}

func NewTaskHandler(taskSvc *task.Service, sessionRepo *session.Repository, llmSettingSvc *auth.LLMSettingService) *TaskHandler {
	return &TaskHandler{
		taskSvc:       taskSvc,
		sessionRepo:   sessionRepo,
		llmSettingSvc: llmSettingSvc,
	}
}

func (h *TaskHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tasks/from-text", h.createFromText)
	rg.GET("/tasks", h.listTasks)
	rg.GET("/tasks/completed", h.listCompletedTasks)
	rg.POST("/tasks", h.createTask)
	rg.GET("/tasks/:id", h.getTask)
	rg.PATCH("/tasks/:id", h.patchTask)
	rg.DELETE("/tasks/:id", h.deleteTask)
	rg.POST("/tasks/:id/steps", h.addStep)
	rg.PATCH("/tasks/:id/steps/:stepId", h.patchStep)
	rg.DELETE("/tasks/:id/steps/:stepId", h.deleteStep)
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

	t, err := h.taskSvc.CreateFromText(c, userID, req.RawText, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取或创建会话
	sess, err := h.sessionRepo.GetTaskSessionOrCreate(c, userID, t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get or create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task":    t,
		"session": sess,
	})
}

// listTasks 获取任务列表
func (h *TaskHandler) listTasks(c *gin.Context) {
	userID := GetUserID(c)
	tasks, err := h.taskSvc.ListTasks(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// listCompletedTasks 获取已完成任务列表
func (h *TaskHandler) listCompletedTasks(c *gin.Context) {
	userID := GetUserID(c)
	tasks, err := h.taskSvc.ListCompletedTasks(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
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

	// 获取或创建会话
	sess, err := h.sessionRepo.GetTaskSessionOrCreate(c, userID, t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get or create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task":    t,
		"session": sess,
	})
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

// deleteTask 删除任务
func (h *TaskHandler) deleteTask(c *gin.Context) {
	userID := GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := h.taskSvc.DeleteTask(c, userID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task deleted successfully"})
}

// patchStep 更新步骤
func (h *TaskHandler) patchStep(c *gin.Context) {
	userID := GetUserID(c)
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	stepIDStr := c.Param("stepId")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
		return
	}

	var fields task.UpdateStepFields
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskSvc.UpdateStep(c, userID, taskID, stepID, fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "step updated"})
}

// addStep 添加步骤
func (h *TaskHandler) addStep(c *gin.Context) {
	userID := GetUserID(c)
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var step task.TaskStep
	if err := c.ShouldBindJSON(&step); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskSvc.AddStep(c, userID, taskID, &step); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"step": step})
}

// deleteStep 删除步骤
func (h *TaskHandler) deleteStep(c *gin.Context) {
	userID := GetUserID(c)
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	stepIDStr := c.Param("stepId")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
		return
	}

	if err := h.taskSvc.DeleteStep(c, userID, taskID, stepID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "step not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "step deleted successfully"})
}
