package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"assistant-qisumi/internal/dependency"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/logger"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	router                 Router
	agents                 map[string]Agent
	taskRepo               *task.Repository
	sessionRepo            *session.Repository
	dependencySvc          *dependency.Service
	db                     *gorm.DB
	llmClient              llm.Client
	chatCompletionsHandler *ChatCompletionsHandler
}

func NewService(
	router Router,
	agents []Agent,
	taskRepo *task.Repository,
	sessionRepo *session.Repository,
	dependencySvc *dependency.Service,
	db *gorm.DB,
	llmClient llm.Client,
) *Service {
	m := make(map[string]Agent)
	for _, ag := range agents {
		m[ag.Name()] = ag
	}

	// 初始化工具执行器映射
	toolMap := NewToolExecutors()

	// 初始化Chat Completions处理器
	chatCompletionsHandler := NewChatCompletionsHandler(llmClient, toolMap)

	// 确保所有使用工具的agent都使用chatCompletionsHandler
	if _, ok := m["executor"]; ok {
		m["executor"] = NewExecutorAgent(llmClient, chatCompletionsHandler)
	}
	if _, ok := m["planner"]; ok {
		m["planner"] = NewPlannerAgent(llmClient, chatCompletionsHandler)
	}
	if _, ok := m["global"]; ok {
		m["global"] = NewGlobalAgent(llmClient, chatCompletionsHandler)
	}

	return &Service{
		router:                 router,
		agents:                 m,
		taskRepo:               taskRepo,
		sessionRepo:            sessionRepo,
		dependencySvc:          dependencySvc,
		db:                     db,
		llmClient:              llmClient,
		chatCompletionsHandler: chatCompletionsHandler,
	}
}

func (s *Service) HandleUserMessage(
	ctx context.Context,
	userID, sessionID uint64,
	userInput string,
	cfg llm.Config,
) (*AgentResponse, error) {
	// 记录请求开始
	logger.Logger.Info("Agent请求开始",
		zap.String("user_id", fmt.Sprintf("%d", userID)),
		zap.String("session_id", fmt.Sprintf("%d", sessionID)),
		zap.String("user_input", userInput),
		zap.String("model", cfg.Model),
		zap.String("base_url", cfg.BaseURL),
	)

	sess, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		logger.Logger.Error("获取会话失败",
			zap.String("session_id", fmt.Sprintf("%d", sessionID)),
			zap.String("error", err.Error()),
		)
		return nil, fmt.Errorf("GetSession failed: %w", err)
	}
	msgs, err := s.sessionRepo.ListRecentMessages(ctx, sessionID, 20)
	if err != nil {
		logger.Logger.Error("获取历史消息失败",
			zap.String("session_id", fmt.Sprintf("%d", sessionID)),
			zap.String("error", err.Error()),
		)
		return nil, fmt.Errorf("ListRecentMessages failed: %w", err)
	}

	var t *task.Task
	if sess.TaskID != nil {
		t, err = s.taskRepo.GetTaskWithSteps(ctx, userID, *sess.TaskID)
		if err != nil {
			return nil, fmt.Errorf("GetTaskWithSteps failed for taskID=%d: %w", *sess.TaskID, err)
		}
	}

	// 获取用户的所有任务（用于全局助手）
	var allTasks []task.Task
	if sess.Type == "global" {
		allTasks, err = s.taskRepo.ListTasks(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("ListTasks failed: %w", err)
		}
	}

	// 获取依赖关系信息（用于Executor判断隐含前置条件）
	var dependencies []task.TaskDependency
	if sess.TaskID != nil {
		dependencies, err = s.taskRepo.GetAllUserDependencies(ctx, userID)
		if err != nil {
			logger.Logger.Warn("获取依赖关系失败，将继续处理",
				zap.String("error", err.Error()),
			)
			// 不中断流程，依赖关系为空时LLM不会执行依赖处理
		}
	}

	req := AgentRequest{
		UserID:       userID,
		Session:      sess,
		Task:         t,
		Tasks:        allTasks,
		Dependencies: dependencies,
		Messages:     msgs,
		UserInput:    userInput,
		Now:          time.Now(),
		LLMConfig:    cfg,
	}

	agentName := s.router.Route(req)
	logger.Logger.Info("路由决策完成",
		zap.String("agent", agentName),
		zap.String("session_type", sess.Type),
		zap.String("session_id", fmt.Sprintf("%d", sessionID)),
	)

	ag, ok := s.agents[agentName]
	if !ok {
		// fallback to executor
		logger.Logger.Warn("Agent未找到，使用fallback",
			zap.String("requested_agent", agentName),
			zap.String("fallback_agent", "executor"),
		)
		ag = s.agents["executor"]
	}

	resp, err := ag.Handle(req)
	if err != nil {
		logger.Logger.Error("Agent处理失败",
			zap.String("agent", agentName),
			zap.String("error", err.Error()),
		)
		return nil, fmt.Errorf("%s agent Handle failed: %w", agentName, err)
	}

	// 首先保存用户消息
	userMsg := session.Message{
		SessionID: sessionID,
		Role:      "user",
		Content:   userInput,
	}
	if err := s.sessionRepo.CreateMessage(ctx, &userMsg); err != nil {
		return nil, fmt.Errorf("CreateMessage (user) failed: %w", err)
	}

	// 1. 开启事务: 应用 TaskPatches 更新 task & steps & dependencies
	if len(resp.TaskPatches) > 0 {
		logger.Logger.Info("开始应用TaskPatches",
			zap.Int("patch_count", len(resp.TaskPatches)),
			zap.String("session_id", fmt.Sprintf("%d", sessionID)),
		)
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 应用TaskPatches
			if err := s.applyTaskPatches(ctx, userID, tx, resp.TaskPatches); err != nil {
				return fmt.Errorf("applyTaskPatches failed: %w", err)
			}
			return nil
		})
		if err != nil {
			logger.Logger.Error("应用TaskPatches失败",
				zap.String("error", err.Error()),
			)
			return nil, fmt.Errorf("transaction failed: %w", err)
		}
		logger.Logger.Info("TaskPatches应用成功",
			zap.Int("patch_count", len(resp.TaskPatches)),
		)
	}

	if strings.TrimSpace(resp.AssistantMessage) == "" {
		resp.AssistantMessage = buildFallbackAssistantMessage(agentName, resp.TaskPatches)
	}

	// 2. 写 assistant 消息: role=assistant, agent_name=ag.Name()
	assistantMsg := session.Message{
		SessionID: sessionID,
		Role:      "assistant",
		AgentName: &agentName,
		Content:   resp.AssistantMessage,
	}
	if err := s.sessionRepo.CreateMessage(ctx, &assistantMsg); err != nil {
		return nil, fmt.Errorf("CreateMessage failed: %w", err)
	}

	logger.Logger.Info("Agent请求完成",
		zap.String("agent", agentName),
		zap.String("session_id", fmt.Sprintf("%d", sessionID)),
		zap.Int("task_patches_count", len(resp.TaskPatches)),
		zap.Int("response_length", len(resp.AssistantMessage)),
	)

	return resp, nil
}

// applyTaskPatches 应用TaskPatches更新数据库
func (s *Service) applyTaskPatches(ctx context.Context, userID uint64, tx *gorm.DB, patches []TaskPatch) error {
	for _, p := range patches {
		switch p.Kind {
		case PatchUpdateTask:
			up := p.UpdateTask
			if up == nil {
				continue
			}
			if err := s.applyUpdateTaskFields(ctx, userID, tx, up.TaskID, up.Fields); err != nil {
				return err
			}

		case PatchUpdateStep:
			up := p.UpdateStep
			if up == nil {
				continue
			}
			if err := s.applyUpdateStepFields(ctx, userID, tx, up.TaskID, up.StepID, up.Fields); err != nil {
				return err
			}

		case PatchAddSteps:
			ap := p.AddSteps
			if ap == nil {
				continue
			}
			if err := s.applyInsertNewSteps(ctx, userID, tx, ap.TaskID, ap.ParentStepID, ap.StepsToInsert); err != nil {
				return err
			}

		case PatchAddDependencies:
			dp := p.AddDependencies
			if dp == nil {
				continue
			}
			if err := s.applyInsertDependencies(ctx, userID, tx, dp.Items); err != nil {
				return err
			}

		case PatchMarkTasksFocusToday:
			fp := p.MarkTasksFocusToday
			if fp == nil {
				continue
			}
			if err := s.applyUpdateTasksFocusToday(ctx, userID, tx, fp.TaskIDs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) applyUpdateTaskFields(ctx context.Context, userID uint64, tx *gorm.DB, taskID uint64, fields task.UpdateTaskFields) error {
	repo := s.taskRepo.WithTx(tx)
	
	// 如果状态变为 done，自动设置 CompletedAt
	if fields.Status != nil && *fields.Status == "done" {
		now := time.Now().Format(time.RFC3339)
		fields.CompletedAt = &now
	} else if fields.Status != nil && *fields.Status != "done" {
		// 如果状态从 done 变为其他状态，清除 CompletedAt
		empty := ""
		fields.CompletedAt = &empty
	}
	
	if err := repo.ApplyUpdateTaskFields(ctx, userID, taskID, fields); err != nil {
		return err
	}

	// 如果状态变为 done，触发依赖处理
	if fields.Status != nil && *fields.Status == "done" {
		if err := s.dependencySvc.OnTaskOrStepDone(ctx, taskID, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) applyUpdateStepFields(ctx context.Context, userID uint64, tx *gorm.DB, taskID, stepID uint64, fields task.UpdateStepFields) error {
	repo := s.taskRepo.WithTx(tx)
	
	// 如果状态变为 done，自动设置 CompletedAt
	if fields.Status != nil && *fields.Status == "done" {
		now := time.Now().Format(time.RFC3339)
		fields.CompletedAt = &now
	} else if fields.Status != nil && *fields.Status != "done" {
		// 如果状态从 done 变为其他状态，清除 CompletedAt
		empty := ""
		fields.CompletedAt = &empty
	}
	
	if err := repo.ApplyUpdateStepFields(ctx, userID, taskID, stepID, fields); err != nil {
		return err
	}

	// 如果步骤状态发生变化，更新任务的 UpdatedAt
	if fields.Status != nil {
		if err := tx.Table("tasks").Where("id = ? AND user_id = ?", taskID, userID).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}
	}

	// 如果状态变为 done，触发依赖处理
	if fields.Status != nil && *fields.Status == "done" {
		if err := s.dependencySvc.OnTaskOrStepDone(ctx, taskID, &stepID); err != nil {
			return err
		}

		// 自动更新任务状态：
		// - 如果有步骤完成，任务状态从 todo 变为 in_progress
		// - 如果所有步骤都完成，任务状态变为 done
		t, err := repo.GetTaskWithSteps(ctx, userID, taskID)
		if err != nil {
			return err
		}

		// 检查所有步骤状态
		allStepsDone := true
		hasSteps := len(t.Steps) > 0
		for _, step := range t.Steps {
			if step.Status != "done" {
				allStepsDone = false
				break
			}
		}

		if hasSteps {
			if allStepsDone {
				// 所有步骤完成，任务状态设为 done
				status := "done"
				now := time.Now().Format(time.RFC3339)
				if err := repo.ApplyUpdateTaskFields(ctx, userID, taskID, task.UpdateTaskFields{
					Status:      &status,
					CompletedAt: &now,
				}); err != nil {
					return err
				}
			} else if t.Status == "todo" {
				// 有步骤完成但未全部完成，任务状态设为 in_progress
				status := "in_progress"
				if err := repo.ApplyUpdateTaskFields(ctx, userID, taskID, task.UpdateTaskFields{
					Status: &status,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Service) applyInsertNewSteps(ctx context.Context, _ uint64, tx *gorm.DB, taskID uint64, _ *uint64, steps []task.NewStepRecord) error {
	repo := s.taskRepo.WithTx(tx)
	var taskSteps []task.TaskStep
	for i, st := range steps {
		ts := task.TaskStep{
			TaskID:      taskID,
			Title:       st.Title,
			Detail:      st.Detail,
			EstimateMin: st.EstimateMinutes,
			OrderIndex:  i, // Simplified
			Status:      "todo",
		}
		taskSteps = append(taskSteps, ts)
	}
	return repo.AddSteps(ctx, taskSteps)
}

func (s *Service) applyInsertDependencies(ctx context.Context, _ uint64, tx *gorm.DB, items []task.DependencyItem) error {
	repo := s.taskRepo.WithTx(tx)
	var deps []task.TaskDependency
	for _, it := range items {
		deps = append(deps, task.TaskDependency{
			PredecessorTaskID: it.PredecessorTaskID,
			PredecessorStepID: it.PredecessorStepID,
			SuccessorTaskID:   it.SuccessorTaskID,
			SuccessorStepID:   it.SuccessorStepID,
			Condition:         it.Condition,
			Action:            it.Action,
		})
	}
	return repo.AddDependencies(ctx, deps)
}

func (s *Service) applyUpdateTasksFocusToday(ctx context.Context, userID uint64, tx *gorm.DB, taskIDs []uint64) error {
	repo := s.taskRepo.WithTx(tx)
	return repo.MarkTasksFocusToday(ctx, userID, taskIDs)
}
