package agent

import (
	"assistant-qisumi/internal/llm"
)

type SummarizerAgent struct {
	llmClient llm.Client
}

func NewSummarizerAgent(llmClient llm.Client) *SummarizerAgent {
	return &SummarizerAgent{llmClient: llmClient}
}

func (a *SummarizerAgent) Name() string { return "summarizer" }

func (a *SummarizerAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// TODO: 构造 messages + tools 调用 LLM
	// TODO: 处理 LLM 响应，生成 assistant 回复文本
	// TODO: 生成对任务的结构化 patch
	return &AgentResponse{
		AssistantMessage: "TODO: summarizer agent reply",
		TaskPatches:      nil,
	}, nil
}
