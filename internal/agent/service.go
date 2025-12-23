package agent

import (
	"context"
	"fmt"
	"time"

	"assistant-qisumi/internal/dependency"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

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
	toolMap := make(map[string]ToolExecutor)
	// 这里可以根据需要添加工具执行器

	// 初始化Chat Completions处理器
	chatCompletionsHandler := NewChatCompletionsHandler(llmClient, toolMap)

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

	sess, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("GetSession failed: %w", err)
	}
	msgs, err := s.sessionRepo.ListRecentMessages(ctx, sessionID, 20)
	if err != nil {
		return nil, fmt.Errorf("ListRecentMessages failed: %w", err)
	}

	var t *task.Task
	if sess.TaskID != nil {
		t, err = s.taskRepo.GetTaskWithSteps(ctx, userID, *sess.TaskID)
		if err != nil {
			return nil, fmt.Errorf("GetTaskWithSteps failed for taskID=%d: %w", *sess.TaskID, err)
		}
	}

	req := AgentRequest{
		UserID:    userID,
		Session:   sess,
		Task:      t,
		Messages:  msgs,
		UserInput: userInput,
		Now:       time.Now(),
		LLMConfig: cfg,
	}

	agentName := s.router.Route(req)
	ag, ok := s.agents[agentName]
	if !ok {
		// fallback to executor
		ag = s.agents["executor"]
	}

	resp, err := ag.Handle(req)
	if err != nil {
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
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 应用TaskPatches
			if err := s.applyTaskPatches(ctx, userID, tx, resp.TaskPatches); err != nil {
				return fmt.Errorf("applyTaskPatches failed: %w", err)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("transaction failed: %w", err)
		}
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
	if err := repo.ApplyUpdateStepFields(ctx, userID, taskID, stepID, fields); err != nil {
		return err
	}

	// 如果状态变为 done，触发依赖处理
	if fields.Status != nil && *fields.Status == "done" {
		if err := s.dependencySvc.OnTaskOrStepDone(ctx, taskID, &stepID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) applyInsertNewSteps(ctx context.Context, userID uint64, tx *gorm.DB, taskID uint64, parentStepID *uint64, steps []task.NewStepRecord) error {
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

func (s *Service) applyInsertDependencies(ctx context.Context, userID uint64, tx *gorm.DB, items []task.DependencyItem) error {
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
