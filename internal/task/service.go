package task

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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
func (s *Service) CreateFromText(ctx context.Context, userID uint64, rawText string, cfg llm.Config) (*Task, error) {
	// 1. 构造 prompt 和 ChatRequest
	prompt := `请将以下文本转换为结构化的任务和步骤。输出格式必须为JSON，包含title、description、due_at、priority、steps数组。
steps数组中的每个元素必须包含title、detail、estimate_minutes（可选）、status默认为"todo"。

文本内容：` + rawText

	messages := []llm.Message{
		{Role: "system", Content: "你是一个任务规划助手，能够将自然语言文本转换为结构化的任务和步骤。"},
		{Role: "user", Content: prompt},
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
		Title       string     `json:"title"`
		Description string     `json:"description"`
		DueAt       *time.Time `json:"due_at,omitempty"`
		Priority    string     `json:"priority"`
		Steps       []Step     `json:"steps"`
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return nil, errors.New("llm returned empty response")
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &taskData); err != nil {
		return nil, err
	}

	// 4. 构造 Task 对象
	t := &Task{
		UserID:      userID,
		Title:       taskData.Title,
		Description: taskData.Description,
		Status:      "todo",
		Priority:    taskData.Priority,
		DueAt:       taskData.DueAt,
		Steps:       taskData.Steps,
	}

	// 5. 调用 repo.InsertTaskWithSteps
	if err := s.repo.InsertTaskWithSteps(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

// ListTasks 获取用户任务列表
func (s *Service) ListTasks(ctx context.Context, userID uint64) ([]Task, error) {
	return s.repo.ListTasks(ctx, userID)
}
