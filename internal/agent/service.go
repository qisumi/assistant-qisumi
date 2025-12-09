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
	router      Router
	agents      map[string]Agent
	taskRepo    *task.Repository
	sessionRepo *session.Repository
	db          *sql.DB
}

func NewService(
	router Router,
	agents []Agent,
	taskRepo *task.Repository,
	sessionRepo *session.Repository,
	db *sql.DB,
) *Service {
	m := make(map[string]Agent)
	for _, ag := range agents {
		m[ag.Name()] = ag
	}
	return &Service{
		router:      router,
		agents:      m,
		taskRepo:    taskRepo,
		sessionRepo: sessionRepo,
		db:          db,
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

		// TODO: 根据TaskPatches类型，调用相应的repo方法更新任务、步骤或依赖关系
		// 这里先简单处理，后续完善

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
