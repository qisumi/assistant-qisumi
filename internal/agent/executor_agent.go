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
	// 1. 构造 messages
	messages, err := BuildExecutorMessages(req.Task, req.Messages, req.UserInput, req.Now)
	if err != nil {
		return nil, err
	}

	// 2. 定义可用工具
	tools := llm.ExecutorTools()

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

	// 2. 处理 LLM 响应，生成 assistant 回复文本和 TaskPatches
	var assistantMessage string
	if len(resp.Choices) > 0 {
		assistantMessage = resp.Choices[0].Message.Content
	}

	taskPatches, err := BuildPatchesFromToolCalls(resp)
	if err != nil {
		// 记录错误但继续返回 assistant 消息
		// log.Printf("build patches error: %v", err)
	}

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
