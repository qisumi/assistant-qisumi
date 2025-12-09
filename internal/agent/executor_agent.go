package agent

import (
	"context"
	"assistant-qisumi/internal/llm"
)

type ExecutorAgent struct {
	llmClient llm.Client
}

func NewExecutorAgent(llmClient llm.Client) *ExecutorAgent {
	return &ExecutorAgent{llmClient: llmClient}
}

func (a *ExecutorAgent) Name() string { return "executor" }

func (a *ExecutorAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 构造 messages + tools 调用 LLM
	messages := []llm.Message{
		{
			Role: "system",
			Content: "你是一个任务执行助手，负责处理用户的任务更新请求。请根据用户输入和当前任务状态，生成相应的回复和任务补丁。",
		},
	}

	// 添加历史消息
	for _, msg := range req.Messages {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前用户输入
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: req.UserInput,
	})

	// 定义可用工具
	tools := []llm.Tool{
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "update_step_status",
				Description: "更新任务步骤状态",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"step_id": map[string]interface{}{
							"type": "integer",
							"description": "步骤ID",
						},
						"status": map[string]interface{}{
							"type": "string",
							"enum": []string{"locked", "todo", "in_progress", "done", "blocked"},
							"description": "新状态",
						},
					},
					"required": []string{"step_id", "status"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "update_task_metadata",
				Description: "更新任务元数据",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type": "string",
							"description": "任务标题",
						},
						"description": map[string]interface{}{
							"type": "string",
							"description": "任务描述",
						},
						"status": map[string]interface{}{
							"type": "string",
							"enum": []string{"todo", "in_progress", "done", "cancelled"},
							"description": "任务状态",
						},
						"priority": map[string]interface{}{
							"type": "string",
							"enum": []string{"low", "medium", "high"},
							"description": "任务优先级",
						},
						"due_at": map[string]interface{}{
							"type": "string",
							"format": "date-time",
							"description": "截止时间",
						},
					},
				},
			},
		},
	}

	// 构造Chat请求
	chatReq := llm.ChatRequest{
		Model:      req.LLMConfig.Model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
	}

	// 调用LLM
	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		return nil, err
	}

	// 2. 处理 LLM 响应，生成 assistant 回复文本
	assistantMessage := ""
	if len(resp.Choices) > 0 {
		assistantMessage = resp.Choices[0].Message.Content
	}

	// 3. 生成对任务的结构化 patch
	// TODO: 解析LLM响应中的tool_calls，生成TaskPatches
	// 这里先返回简单的响应，后续完善tool_calls解析

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      []TaskPatch{},
	}, nil
}
