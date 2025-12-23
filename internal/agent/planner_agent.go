package agent

import (
	"context"

	"assistant-qisumi/internal/llm"
)

type PlannerAgent struct {
	llmClient              llm.Client
	chatCompletionsHandler *ChatCompletionsHandler
}

func NewPlannerAgent(llmClient llm.Client, chatCompletionsHandler *ChatCompletionsHandler) *PlannerAgent {
	return &PlannerAgent{
		llmClient:              llmClient,
		chatCompletionsHandler: chatCompletionsHandler,
	}
}

func (a *PlannerAgent) Name() string { return "planner" }

func (a *PlannerAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 构造 messages
	messages, err := BuildPlannerMessages(req.Task, req.Messages, req.UserInput, req.Now)
	if err != nil {
		return nil, err
	}

	// 2. 准备tools
	tools := llm.PlannerTools()

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
