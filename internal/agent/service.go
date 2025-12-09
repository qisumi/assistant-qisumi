package agent

import (
	"context"
	"database/sql"
	"time"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"
)

type Service struct {
	router                 Router
	agents                 map[string]Agent
	taskRepo               *task.Repository
	sessionRepo            *session.Repository
	db                     *sql.DB
	llmClient              llm.Client
	chatCompletionsHandler *ChatCompletionsHandler
}

func NewService(
	router Router,
	agents []Agent,
	taskRepo *task.Repository,
	sessionRepo *session.Repository,
	db *sql.DB,
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
		return nil, err
	}
	msgs, err := s.sessionRepo.ListRecentMessages(ctx, sessionID, 20)
	if err != nil {
		return nil, err
	}

	var t *task.Task
	if sess.TaskID != nil {
		t, err = s.taskRepo.GetTaskWithSteps(ctx, userID, *sess.TaskID)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	// 1. 开启事务: 应用 TaskPatches 更新 task & steps & dependencies
	if len(resp.TaskPatches) > 0 {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		// 应用TaskPatches
		if err := s.applyTaskPatches(ctx, tx, resp.TaskPatches); err != nil {
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, err
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
		return nil, err
	}

	return resp, nil
}

// applyTaskPatches 应用TaskPatches更新数据库
func (s *Service) applyTaskPatches(ctx context.Context, tx *sql.Tx, patches []TaskPatch) error {
	for _, patch := range patches {
		switch patch.Type {
		case "update_task":
			if err := s.applyUpdateTaskPatch(ctx, tx, patch); err != nil {
				return err
			}
		case "update_step":
			if err := s.applyUpdateStepPatch(ctx, tx, patch); err != nil {
				return err
			}
		case "add_steps":
			if err := s.applyAddStepsPatch(ctx, tx, patch); err != nil {
				return err
			}
		case "add_dependencies":
			if err := s.applyAddDependenciesPatch(ctx, tx, patch); err != nil {
				return err
			}
		case "mark_tasks_focus_today":
			if err := s.applyMarkTasksFocusTodayPatch(ctx, tx, patch); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyUpdateTaskPatch 应用更新任务的TaskPatch
func (s *Service) applyUpdateTaskPatch(ctx context.Context, tx *sql.Tx, patch TaskPatch) error {
	taskID, ok := patch.Payload["task_id"].(uint64)
	if !ok {
		return nil
	}

	fields, ok := patch.Payload["fields"].(map[string]interface{})
	if !ok {
		return nil
	}

	// 调用taskRepo的更新方法
	return s.taskRepo.UpdateTask(ctx, tx, taskID, fields)
}

// applyUpdateStepPatch 应用更新步骤的TaskPatch
func (s *Service) applyUpdateStepPatch(ctx context.Context, tx *sql.Tx, patch TaskPatch) error {
	taskID, ok := patch.Payload["task_id"].(uint64)
	if !ok {
		return nil
	}

	stepID, ok := patch.Payload["step_id"].(uint64)
	if !ok {
		return nil
	}

	fields, ok := patch.Payload["fields"].(map[string]interface{})
	if !ok {
		return nil
	}

	// 调用taskRepo的更新步骤方法
	return s.taskRepo.UpdateStep(ctx, tx, taskID, stepID, fields)
}

// applyAddStepsPatch 应用添加步骤的TaskPatch
func (s *Service) applyAddStepsPatch(ctx context.Context, tx *sql.Tx, patch TaskPatch) error {
	// 从payload中获取任务ID和步骤信息
	taskID, ok := patch.Payload["task_id"].(uint64)
	if !ok {
		return nil
	}

	stepsData, ok := patch.Payload["steps"].([]interface{})
	if !ok {
		return nil
	}

	// 转换为task.Step类型
	var steps []task.Step
	for _, stepData := range stepsData {
		stepMap, ok := stepData.(map[string]interface{})
		if !ok {
			continue
		}

		// 解析基本字段
		title, _ := stepMap["title"].(string)
		detail, _ := stepMap["detail"].(string)
		status, _ := stepMap["status"].(string)
		if status == "" {
			status = "todo"
		}

		// 解析估计时间
		estimateMinutes := 30 // 默认30分钟
		if estMin, ok := stepMap["estimate_minutes"].(float64); ok {
			estimateMinutes = int(estMin)
		}
		estMinPtr := &estimateMinutes

		// 解析顺序索引
		orderIndex := 0
		if ordIdx, ok := stepMap["order_index"].(float64); ok {
			orderIndex = int(ordIdx)
		}

		// 解析计划时间
		var plannedStart, plannedEnd *time.Time
		if ps, ok := stepMap["planned_start"].(string); ok && ps != "" {
			if t, err := time.Parse(time.RFC3339, ps); err == nil {
				plannedStart = &t
			}
		}
		if pe, ok := stepMap["planned_end"].(string); ok && pe != "" {
			if t, err := time.Parse(time.RFC3339, pe); err == nil {
				plannedEnd = &t
			}
		}

		// 创建步骤对象
		step := task.Step{
			TaskID:       taskID,
			OrderIndex:   orderIndex,
			Title:        title,
			Detail:       detail,
			Status:       status,
			EstimateMin:  estMinPtr,
			PlannedStart: plannedStart,
			PlannedEnd:   plannedEnd,
		}

		steps = append(steps, step)
	}

	// 调用taskRepo的添加步骤方法
	return s.taskRepo.AddSteps(ctx, tx, steps)
}

// applyAddDependenciesPatch 应用添加依赖的TaskPatch
func (s *Service) applyAddDependenciesPatch(ctx context.Context, tx *sql.Tx, patch TaskPatch) error {
	// 从payload中获取依赖项信息
	items, ok := patch.Payload["items"].([]interface{})
	if !ok {
		return nil
	}

	// 转换为task.Dependency类型
	var dependencies []task.Dependency
	for _, itemData := range items {
		itemMap, ok := itemData.(map[string]interface{})
		if !ok {
			continue
		}

		// 解析前置任务/步骤ID
		predecessorTaskID, ok := itemMap["predecessor_task_id"].(float64)
		if !ok {
			continue
		}

		// 解析前置步骤ID（可选）
		var predecessorStepID *uint64
		if psID, ok := itemMap["predecessor_step_id"].(float64); ok {
			id := uint64(psID)
			predecessorStepID = &id
		}

		// 解析后置任务/步骤ID
		successorTaskID, ok := itemMap["successor_task_id"].(float64)
		if !ok {
			continue
		}

		// 解析后置步骤ID（可选）
		var successorStepID *uint64
		if ssID, ok := itemMap["successor_step_id"].(float64); ok {
			id := uint64(ssID)
			successorStepID = &id
		}

		// 解析条件和动作
		condition, _ := itemMap["condition"].(string)
		action, _ := itemMap["action"].(string)

		// 创建依赖对象
		dep := task.Dependency{
			PredecessorTaskID:   uint64(predecessorTaskID),
			PredecessorStepID:   predecessorStepID,
			SuccessorTaskID:     uint64(successorTaskID),
			SuccessorStepID:     successorStepID,
			DependencyCondition: condition,
			Action:              action,
		}

		dependencies = append(dependencies, dep)
	}

	// 调用taskRepo的添加依赖方法
	return s.taskRepo.AddDependencies(ctx, tx, dependencies)
}

// applyMarkTasksFocusTodayPatch 应用标记今日重点任务的TaskPatch
func (s *Service) applyMarkTasksFocusTodayPatch(ctx context.Context, tx *sql.Tx, patch TaskPatch) error {
	// 从payload中获取用户ID和任务ID列表
	userID, ok := patch.Payload["user_id"].(uint64)
	if !ok {
		return nil
	}

	taskIDsData, ok := patch.Payload["task_ids"].([]interface{})
	if !ok {
		return nil
	}

	// 转换为[]uint64类型
	var taskIDs []uint64
	for _, taskIDData := range taskIDsData {
		if taskID, ok := taskIDData.(float64); ok {
			taskIDs = append(taskIDs, uint64(taskID))
		}
	}

	// 调用taskRepo的标记方法
	return s.taskRepo.MarkTasksFocusToday(ctx, tx, userID, taskIDs)
}
