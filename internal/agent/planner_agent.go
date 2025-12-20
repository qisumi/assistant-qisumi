package agent

import (
	"assistant-qisumi/internal/llm"
	"context"
	"fmt"
)

type PlannerAgent struct {
	llmClient llm.Client
}

func NewPlannerAgent(llmClient llm.Client) *PlannerAgent {
	return &PlannerAgent{llmClient: llmClient}
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

	// 3. 构造Chat请求
	chatReq := llm.ChatRequest{
		Model:      req.LLMConfig.Model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
	}

	// 4. 调用LLM
	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// 5. 处理LLM响应
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	chatMsg := resp.Choices[0].Message
	assistantMessage := chatMsg.Content

	taskPatches, err := BuildPatchesFromToolCalls(resp)
	if err != nil {
		// 记录错误但继续返回 assistant 消息
	}

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
