package agent

import (
	"context"

	"assistant-qisumi/internal/llm"
)

type ExecutorAgent struct {
	llmClient              llm.Client
	chatCompletionsHandler *ChatCompletionsHandler
}

func NewExecutorAgent(llmClient llm.Client, chatCompletionsHandler *ChatCompletionsHandler) *ExecutorAgent {
	return &ExecutorAgent{
		llmClient:              llmClient,
		chatCompletionsHandler: chatCompletionsHandler,
	}
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

	// 3. 使用 ChatCompletionsHandler 处理完整的工具调用流程
	// 这会自动处理：初始LLM调用 -> 工具执行 -> 二次LLM调用生成最终回复
	assistantMessage, taskPatches, err := a.chatCompletionsHandler.HandleChatCompletions(
		context.Background(),
		req.LLMConfig,
		messages,
		tools,
	)
	if err != nil {
		return nil, err
	}

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
