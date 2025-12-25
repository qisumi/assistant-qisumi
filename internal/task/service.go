package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"assistant-qisumi/internal/common"
	"assistant-qisumi/internal/llm"
)

type Service struct {
	repo      *Repository
	llmClient llm.Client
}

func NewService(repo *Repository, llmClient llm.Client) *Service {
	return &Service{repo: repo, llmClient: llmClient}
}

// CreateFromText: 调用 LLM 把一段文本变成 Task + Steps
// 使用 TaskCreationAgent 的 prompt 来生成高质量的任务和步骤
func (s *Service) CreateFromText(ctx context.Context, userID uint64, rawText string, cfg llm.Config) (*Task, error) {
	// 1. 构造 messages（使用 TaskCreationSystemPrompt）
	messages := []llm.Message{
		{
			Role:    "system",
			Content: common.TaskCreationSystemPrompt,
		},
		{
			Role:    "system",
			Content: "当前时间 now: " + time.Now().Format(time.RFC3339),
		},
		{
			Role:    "user",
			Content: rawText,
		},
	}

	req := llm.ChatRequest{
		Model:    cfg.Model,
		Messages: messages,
	}

	// 2. 调用 llmClient.Chat
	resp, err := s.llmClient.Chat(ctx, cfg, req)
	if err != nil {
		return nil, err
	}

	// 3. 解析 JSON -> Task + Steps
	var taskData struct {
		Title       string             `json:"title"`
		Description string             `json:"description"`
		DueAt       *FlexibleTime      `json:"due_at,omitempty"`
		Priority    string             `json:"priority"`
		Steps       []taskCreationStep `json:"steps"`
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return nil, errors.New("llm returned empty response")
	}

	content := common.ExtractJSON(resp.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(content), &taskData); err != nil {
		return nil, fmt.Errorf("failed to parse llm response: %w, content: %s", err, content)
	}

	// 4. 构造 Task 对象
	steps := make([]TaskStep, len(taskData.Steps))
	for i, s := range taskData.Steps {
		steps[i] = TaskStep{
			Title:       s.Title,
			Detail:      s.Detail,
			EstimateMin: &s.EstimateMinutes,
			OrderIndex:  i,
			Status:      "todo",
		}
	}

	t := &Task{
		UserID:      userID,
		Title:       taskData.Title,
		Description: taskData.Description,
		Status:      "todo",
		Priority:    taskData.Priority,
		DueAt:       taskData.DueAt,
		Steps:       steps,
	}

	// 5. 调用 repo.InsertTaskWithSteps
	if err := s.repo.InsertTaskWithSteps(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

// taskCreationStep 用于解析 TaskCreationAgent 生成的步骤数据
type taskCreationStep struct {
	Title           string `json:"title"`
	Detail          string `json:"detail"`
	EstimateMinutes int    `json:"estimate_minutes"`
	OrderIndex      int    `json:"order_index"`
}

// ListTasks 获取用户任务列表
func (s *Service) ListTasks(ctx context.Context, userID uint64) ([]Task, error) {
	return s.repo.ListTasks(ctx, userID)
}

// ListCompletedTasks 获取用户已完成的任务列表
func (s *Service) ListCompletedTasks(ctx context.Context, userID uint64) ([]Task, error) {
	return s.repo.ListCompletedTasks(ctx, userID)
}

// GetTask 获取任务详情
func (s *Service) GetTask(ctx context.Context, userID, taskID uint64) (*Task, error) {
	return s.repo.GetTaskWithSteps(ctx, userID, taskID)
}

// UpdateTask 更新任务
func (s *Service) UpdateTask(ctx context.Context, userID, taskID uint64, fields UpdateTaskFields) error {
	return s.repo.ApplyUpdateTaskFields(ctx, userID, taskID, fields)
}

// UpdateStep 更新步骤
func (s *Service) UpdateStep(ctx context.Context, userID, taskID, stepID uint64, fields UpdateStepFields) error {
	if err := s.repo.ApplyUpdateStepFields(ctx, userID, taskID, stepID, fields); err != nil {
		return err
	}
	// 更新步骤后，总是更新任务的 updated_at
	return s.repo.db.WithContext(ctx).Table("tasks").Where("id = ? AND user_id = ?", taskID, userID).Update("updated_at", s.repo.db.NowFunc()).Error
}

// CreateTask 创建任务
func (s *Service) CreateTask(ctx context.Context, t *Task) error {
	return s.repo.InsertTaskWithSteps(ctx, t)
}

// DeleteTask 删除任务
func (s *Service) DeleteTask(ctx context.Context, userID, taskID uint64) error {
	// 验证任务是否存在且属于该用户
	_, err := s.repo.GetTaskWithSteps(ctx, userID, taskID)
	if err != nil {
		return err
	}

	// 执行删除
	return s.repo.DeleteTask(ctx, userID, taskID)
}

// AddStep 添加步骤
func (s *Service) AddStep(ctx context.Context, userID, taskID uint64, step *TaskStep) error {
	// 验证任务是否存在且属于该用户
	_, err := s.repo.GetTaskWithSteps(ctx, userID, taskID)
	if err != nil {
		return err
	}

	step.TaskID = taskID
	return s.repo.AddStep(ctx, step)
}

// DeleteStep 删除步骤
func (s *Service) DeleteStep(ctx context.Context, userID, taskID, stepID uint64) error {
	return s.repo.DeleteStep(ctx, userID, taskID, stepID)
}
