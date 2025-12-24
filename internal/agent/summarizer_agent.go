package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/logger"

	"go.uber.org/zap"
)

type SummarizerAgent struct {
	llmClient llm.Client
}

func NewSummarizerAgent(llmClient llm.Client) *SummarizerAgent {
	return &SummarizerAgent{llmClient: llmClient}
}

func (a *SummarizerAgent) Name() string { return "summarizer" }

func (a *SummarizerAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	logger.Logger.Info("SummarizerAgent开始处理",
		zap.String("session_id", fmt.Sprintf("%d", req.Session.ID)),
		zap.String("user_input", req.UserInput),
		zap.Int("history_messages", len(req.Messages)),
	)

	// 1. 构造 messages 调用 LLM
	messages := []llm.Message{
		{
			Role: "system",
			Content: `你叫小奇，是用户的助手兼秘书，风格「严谨但有人情味」。
你的身份是一个任务总结助手（Summarizer Agent）。

你收到的是：
- 当前任务的结构化信息（task 和 steps）
- 该任务的最近若干条对话消息

你的职责：
1. 总结这个任务的当前状态，包括：
   - 总共有多少个步骤，已完成/未完成数量
   - 关键的完成里程碑
   - 是否即将到期或已经逾期
2. 总结最近的对话，看有没有：
   - 用户已经做了什么决策
   - AI 之前给过的建议（可以简要复述）
3. 给出简洁的自然语言反馈，可以包含：
   - 任务进度概览
   - 一两条下一步行动建议

注意：
- 不需要调用任何工具，也不修改任务状态。
- 重点是「解释当前状况」而不是「重排计划」。
- 输出语言尽量简洁友好，避免啰嗦。
- 面向用户的回复里不要展示 task_id/step_id 等内部编号；用「任务标题/步骤标题/序号」来表达即可（除非用户明确要求看编号）。`,
		},
	}

	// 添加当前任务状态信息
	if req.Task != nil {
		if taskJSON, err := json.Marshal(req.Task); err == nil {
			messages = append(messages, llm.Message{
				Role:    "system",
				Content: "当前任务状态（只读 JSON）：\n" + string(taskJSON),
			})
			logger.Logger.Debug("添加任务状态信息",
				zap.Uint64("task_id", req.Task.ID),
			)
		}
	}

	// 添加当前时间信息
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: "当前时间 now: " + req.Now.Format(time.RFC3339),
	})

	// 添加历史消息
	messages = append(messages, historyToLLMMessages(req.Messages)...)
	logger.Logger.Debug("添加历史消息",
		zap.Int("history_count", len(req.Messages)),
	)

	// 添加当前用户输入
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: req.UserInput,
	})

	// SummarizerAgent 不需要工具，只做总结
	tools := []llm.Tool{}

	// 构造Chat请求
	chatReq := llm.ChatRequest{
		Model:      req.LLMConfig.Model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "none", // 明确不需要工具调用
	}

	// 调用LLM
	logger.Logger.Debug("发送Summarizer请求到LLM",
		zap.String("model", req.LLMConfig.Model),
	)
	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		logger.Logger.Error("Summarizer LLM调用失败",
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	// 2. 处理 LLM 响应，生成 assistant 回复文本
	var assistantMessage string

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		assistantMessage = choice.Message.Content
		logger.Logger.Debug("收到Summarizer响应",
			zap.Int("response_length", len(assistantMessage)),
		)
	}

	logger.Logger.Info("SummarizerAgent处理完成",
		zap.Int("response_length", len(assistantMessage)),
	)

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      []TaskPatch{}, // SummarizerAgent 不产生 TaskPatches
	}, nil
}
