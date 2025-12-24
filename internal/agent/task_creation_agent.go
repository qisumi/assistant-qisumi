package agent

import (
	"context"
	"encoding/json"
	"time"

	"assistant-qisumi/internal/llm"
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
			Role: "system",
			Content: `你是一个任务规划助手（Task Creation Agent）。

用户会提供一段自然语言文本，它可能是：
- 一次会议纪要
- 一段聊天记录
- 一个自己写的备忘录
- 一段目标描述（例如"本周完成 AIGC 小论文"）

你的目标是：
1. 从这段文本中抽取出一个「任务」（task）及其基本信息：
   - title: 任务标题，用一句话概括
   - description: 简短描述
   - due_at: 任务截止时间（ISO 8601 格式字符串，例如 2025-12-08T23:00:00；如果文本没有明确时间，可以为 null）
   - priority: low / medium / high，基于文本紧急程度和重要性进行判断
2. 把任务拆解为一个有顺序的步骤列表 steps：
   - 每个步骤包含：
     - title: 步骤标题
     - detail: 说明
     - estimate_minutes: 预估需要的分钟数（可以粗略估计）
     - order_index: 从 1 开始的整数，代表执行顺序

请严格输出一个 JSON 对象，字段必须为：
{
  "title": "...",
  "description": "...",
  "due_at": "..." or null,
  "priority": "low|medium|high",
  "steps": [
    {
      "title": "...",
      "detail": "...",
      "estimate_minutes": 60,
      "order_index": 1
    }
  ]
}

不要输出任何多余的文本或注释，不要加 Markdown，只返回 JSON。
如果文本里面包含多个大任务，你可以倾向于专注于最大的核心任务，并把其余内容融入 description 或 steps 中。`,
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
