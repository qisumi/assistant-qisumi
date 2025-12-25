package agent

import (
	"context"
	"time"

	"assistant-qisumi/internal/domain"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/prompts"
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

	// 2. 处理 LLM 响应
	if len(resp.Choices) == 0 {
		return &AgentResponse{
			AssistantMessage: "未能从文本生成任务，请重试。",
			TaskPatches:      []TaskPatch{},
		}, nil
	}

	// 3. 使用 domain 包的共享逻辑解析 JSON 响应
	output, err := domain.ParseTaskCreationResponse(resp.Choices[0].Message.Content)
	if err != nil {
		return &AgentResponse{
			AssistantMessage: "未能解析生成的任务数据，请重试。",
			TaskPatches:      []TaskPatch{},
		}, nil
	}

	// 4. 生成 TaskPatches
	patches := []TaskPatch{
		{
			Kind: PatchCreateTask,
			CreateTask: &CreateTaskPatch{
				Title:       output.Title,
				Description: output.Description,
				DueAt:       output.DueAtString(),
				Priority:    output.Priority,
				Steps:       output.ToNewStepRecords(),
			},
		},
	}

	return &AgentResponse{
		AssistantMessage: "好的，我已经把这段内容整理成一个任务，并拆成了可执行的步骤。",
		TaskPatches:      patches,
	}, nil
}
