package agent

import (
	"assistant-qisumi/internal/llm"
)

type PlannerAgent struct {
	llmClient llm.Client
}

func NewPlannerAgent(llmClient llm.Client) *PlannerAgent {
	return &PlannerAgent{llmClient: llmClient}
}

func (a *PlannerAgent) Name() string { return "planner" }

func (a *PlannerAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// TODO: 构造 messages + tools 调用 LLM
	// TODO: 处理 LLM 响应，生成 assistant 回复文本
	// TODO: 生成对任务的结构化 patch
	return &AgentResponse{
		AssistantMessage: "TODO: planner agent reply",
		TaskPatches:      nil,
	}, nil
}
