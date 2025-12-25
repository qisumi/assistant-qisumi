package agent

import (
	"context"
	"encoding/json"
	"time"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/prompts"
	"assistant-qisumi/internal/task"
)

type TaskCreationAgent struct {
	llmClient llm.Client
}

func NewTaskCreationAgent(llmClient llm.Client) *TaskCreationAgent {
	return &TaskCreationAgent{llmClient: llmClient}
}

func (a *TaskCreationAgent) Name() string { return "task_creation" }

func (a *TaskCreationAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 构造 messages 调用 LLM
	messages := []llm.Message{
		{
			Role:    "system",
			Content: prompts.TaskCreationSystemPrompt,
		},
		{
			Role:    "system",
			Content: "当前时间 now: " + req.Now.Format(time.RFC3339),
		},
		{
			Role:    "user",
			Content: req.UserInput,
		},
	}

	// 构造Chat请求
	chatReq := llm.ChatRequest{
		Model:    req.LLMConfig.Model,
		Messages: messages,
	}

	// 调用LLM
	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		return nil, err
	}

	// 2. 处理 LLM 响应，生成 assistant 回复文本和 TaskPatches
	if len(resp.Choices) == 0 {
		return &AgentResponse{
			AssistantMessage: "未能从文本生成任务，请重试。",
			TaskPatches:      []TaskPatch{},
		}, nil
	}

	// 3. 解析JSON响应
	var taskData struct {
		Title       string             `json:"title"`
		Description string             `json:"description"`
		DueAt       *task.FlexibleTime `json:"due_at,omitempty"`
		Priority    string             `json:"priority"`
		Steps       []taskCreationStep `json:"steps"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &taskData); err != nil {
		return &AgentResponse{
			AssistantMessage: "未能解析生成的任务数据，请重试。",
			TaskPatches:      []TaskPatch{},
		}, nil
	}

	// 4. 生成TaskPatches
	var dueAtStr *string
	if taskData.DueAt != nil {
		s := taskData.DueAt.ToTime().Format(time.RFC3339)
		dueAtStr = &s
	}

	records := make([]task.NewStepRecord, 0, len(taskData.Steps))
	for _, s := range taskData.Steps {
		est := s.EstimateMinutes
		records = append(records, task.NewStepRecord{
			Title:           s.Title,
			Detail:          s.Detail,
			EstimateMinutes: &est,
		})
	}

	patches := []TaskPatch{
		{
			Kind: PatchCreateTask,
			CreateTask: &CreateTaskPatch{
				Title:       taskData.Title,
				Description: taskData.Description,
				DueAt:       dueAtStr,
				Priority:    taskData.Priority,
				Steps:       records,
			},
		},
	}

	return &AgentResponse{
		AssistantMessage: "好的，我已经把这段内容整理成一个任务，并拆成了可执行的步骤。",
		TaskPatches:      patches,
	}, nil
}

// 用于解析TaskCreationAgent生成的步骤数据
type taskCreationStep struct {
	Title           string `json:"title"`
	Detail          string `json:"detail"`
	EstimateMinutes int    `json:"estimate_minutes"`
	OrderIndex      int    `json:"order_index"`
}
