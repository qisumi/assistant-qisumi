package http

import (
	"errors"

	"assistant-qisumi/internal/auth"
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
		R.BadRequest(c, err.Error())
		return
	}

	cfg, err := GetLLMConfig(c, h.llmSettingSvc, userID)
	if err != nil {
		return
	}

	t, err := h.taskSvc.CreateFromText(c, userID, req.RawText, *cfg)
	if err != nil {
		R.InternalError(c, err.Error())
		return
	}

	// 获取或创建会话
	sess, err := h.sessionRepo.GetTaskSessionOrCreate(c, userID, t.ID)
	if err != nil {
		R.InternalError(c, "failed to get or create session")
		return
	}

	R.Success(c, gin.H{
		"task":    t,
		"session": sess,
	})
}

// listTasks 获取任务列表
func (h *TaskHandler) listTasks(c *gin.Context) {
	userID := GetUserID(c)
	tasks, err := h.taskSvc.ListTasks(c, userID)
	if err != nil {
		R.InternalError(c, err.Error())
		return
	}
	R.Success(c, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// listCompletedTasks 获取已完成任务列表
func (h *TaskHandler) listCompletedTasks(c *gin.Context) {
	userID := GetUserID(c)
	tasks, err := h.taskSvc.ListCompletedTasks(c, userID)
	if err != nil {
		R.InternalError(c, err.Error())
		return
	}
	R.Success(c, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// getTask 获取任务详情
func (h *TaskHandler) getTask(c *gin.Context) {
	userID := GetUserID(c)
	id, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	t, err := h.taskSvc.GetTask(c, userID, id)
	if err != nil {
		R.InternalError(c, err.Error())
		return
	}

	// 获取或创建会话
	sess, err := h.sessionRepo.GetTaskSessionOrCreate(c, userID, t.ID)
	if err != nil {
		R.InternalError(c, "failed to get or create session")
		return
	}

	R.Success(c, gin.H{
		"task":    t,
		"session": sess,
	})
}

// patchTask 更新任务
func (h *TaskHandler) patchTask(c *gin.Context) {
	userID := GetUserID(c)
	id, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	var fields task.UpdateTaskFields
	if err := c.ShouldBindJSON(&fields); err != nil {
		R.BadRequest(c, err.Error())
		return
	}

	if err := h.taskSvc.UpdateTask(c, userID, id, fields); err != nil {
		R.InternalError(c, err.Error())
		return
	}
	R.SuccessWithMessage(c, "task updated", nil)
}

// createTask 创建任务
func (h *TaskHandler) createTask(c *gin.Context) {
	userID := GetUserID(c)
	var t task.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		R.BadRequest(c, err.Error())
		return
	}
	t.UserID = userID

	if err := h.taskSvc.CreateTask(c, &t); err != nil {
		R.InternalError(c, err.Error())
		return
	}
	R.Success(c, gin.H{"task": t})
}

// deleteTask 删除任务
func (h *TaskHandler) deleteTask(c *gin.Context) {
	userID := GetUserID(c)
	id, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	if err := h.taskSvc.DeleteTask(c, userID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			R.NotFound(c, "task not found")
		} else {
			R.InternalError(c, err.Error())
		}
		return
	}

	R.SuccessWithMessage(c, "task deleted successfully", nil)
}

// patchStep 更新步骤
func (h *TaskHandler) patchStep(c *gin.Context) {
	userID := GetUserID(c)
	taskID, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	stepID, err := ParseUint64Param(c, "stepId")
	if err != nil {
		return
	}

	var fields task.UpdateStepFields
	if err := c.ShouldBindJSON(&fields); err != nil {
		R.BadRequest(c, err.Error())
		return
	}

	if err := h.taskSvc.UpdateStep(c, userID, taskID, stepID, fields); err != nil {
		R.InternalError(c, err.Error())
		return
	}
	R.SuccessWithMessage(c, "step updated", nil)
}

// addStep 添加步骤
func (h *TaskHandler) addStep(c *gin.Context) {
	userID := GetUserID(c)
	taskID, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	var step task.TaskStep
	if err := c.ShouldBindJSON(&step); err != nil {
		R.BadRequest(c, err.Error())
		return
	}

	if err := h.taskSvc.AddStep(c, userID, taskID, &step); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			R.NotFound(c, "task not found")
		} else {
			R.InternalError(c, err.Error())
		}
		return
	}

	R.Success(c, gin.H{"step": step})
}

// deleteStep 删除步骤
func (h *TaskHandler) deleteStep(c *gin.Context) {
	userID := GetUserID(c)
	taskID, err := ParseUint64Param(c, "id")
	if err != nil {
		return
	}

	stepID, err := ParseUint64Param(c, "stepId")
	if err != nil {
		return
	}

	if err := h.taskSvc.DeleteStep(c, userID, taskID, stepID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			R.NotFound(c, "step not found")
		} else {
			R.InternalError(c, err.Error())
		}
		return
	}

	R.SuccessWithMessage(c, "step deleted successfully", nil)
}
