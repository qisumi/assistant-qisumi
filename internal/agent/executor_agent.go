package agent

import (
	"context"
	"fmt"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/logger"

	"go.uber.org/zap"
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
	logger.Logger.Info("ExecutorAgent开始处理",
		zap.String("session_id", fmt.Sprintf("%d", req.Session.ID)),
		zap.String("user_input", req.UserInput),
		zap.Int("history_messages", len(req.Messages)),
	)

	// 1. 构造 messages
	messages, err := BuildExecutorMessages(req.Task, req.Dependencies, req.Messages, req.UserInput, req.Now)
	if err != nil {
		logger.Logger.Error("构造Executor消息失败",
			zap.String("error", err.Error()),
		)
		return nil, err
	}
	logger.Logger.Debug("Executor消息构造完成",
		zap.Int("messages_count", len(messages)),
	)

	// 2. 定义可用工具
	tools := llm.ExecutorTools()
	logger.Logger.Debug("Executor工具定义",
		zap.Int("tools_count", len(tools)),
	)

	// 3. 使用 ChatCompletionsHandler 处理完整的工具调用流程
	// 这会自动处理：初始LLM调用 -> 工具执行 -> 二次LLM调用生成最终回复
	assistantMessage, taskPatches, err := a.chatCompletionsHandler.HandleChatCompletions(
		context.Background(),
		req.LLMConfig,
		messages,
		tools,
	)
	if err != nil {
		logger.Logger.Error("Executor处理失败",
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Logger.Info("ExecutorAgent处理完成",
		zap.Int("task_patches_count", len(taskPatches)),
		zap.Int("response_length", len(assistantMessage)),
	)

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
