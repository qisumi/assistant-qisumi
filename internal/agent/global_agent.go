package agent

import (
	"assistant-qisumi/internal/llm"
)

type GlobalAgent struct {
	llmClient llm.Client
}

func NewGlobalAgent(llmClient llm.Client) *GlobalAgent {
	return &GlobalAgent{llmClient: llmClient}
}

func (a *GlobalAgent) Name() string { return "global" }

func (a *GlobalAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// TODO: 构造 messages + tools 调用 LLM
	// TODO: 处理 LLM 响应，生成 assistant 回复文本
	// TODO: 生成对任务的结构化 patch
	return &AgentResponse{
		AssistantMessage: "TODO: global agent reply",
		TaskPatches:      nil,
	}, nil
}
