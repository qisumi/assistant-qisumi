package test

import (
	"assistant-qisumi/internal/agent"
	"assistant-qisumi/internal/llm"
	"context"
)

type mockLLMClient struct{}

func (m *mockLLMClient) Chat(ctx context.Context, cfg llm.Config, req llm.ChatRequest) (*llm.ChatResponse, error) {
	resp := &llm.ChatResponse{
		Choices: []struct {
			Message      llm.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{
			{
				Message: llm.ChatMessage{
					Role:    "assistant",
					Content: `{"title":"Test Task","description":"Test Description","priority":"high","steps":[{"title":"Step 1","detail":"Detail 1"}]}`,
				},
			},
		},
	}
	return resp, nil
}

type mockAgent struct{}

func (m *mockAgent) Name() string { return "executor" }
func (m *mockAgent) Handle(req agent.AgentRequest) (*agent.AgentResponse, error) {
	return &agent.AgentResponse{AssistantMessage: "Hello from agent"}, nil
}

type mockRouter struct{}

func (m *mockRouter) Route(req agent.AgentRequest) string { return "executor" }
