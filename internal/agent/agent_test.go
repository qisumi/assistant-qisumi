package agent

import (
	"context"
	"testing"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
)

// MockLLMClient 用于测试的LLM客户端模拟
type MockLLMClient struct{}

func (m *MockLLMClient) Chat(ctx context.Context, cfg llm.Config, req llm.ChatRequest) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{
		Choices: []struct {
			Message      llm.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{
			{
				Message: llm.ChatMessage{Content: "{\"agent\": \"executor\"}"},
			},
		},
	}, nil
}

// TestSimpleRouter 测试简单路由功能
func TestSimpleRouter(t *testing.T) {
	router := NewSimpleRouter()

	tests := []struct {
		name     string
		req      AgentRequest
		expected string
	}{
		{
			name: "global session",
			req: AgentRequest{
				Session: &session.Session{Type: "global"},
			},
			expected: "global",
		},
		{
			name: "summarizer request",
			req: AgentRequest{
				Session:   &session.Session{Type: "task"},
				UserInput: "总结一下这个任务的进度",
			},
			expected: "summarizer",
		},
		{
			name: "planner request",
			req: AgentRequest{
				Session:   &session.Session{Type: "task"},
				UserInput: "重新规划一下任务",
			},
			expected: "planner",
		},
		{
			name: "executor request",
			req: AgentRequest{
				Session:   &session.Session{Type: "task"},
				UserInput: "我完成了第一步",
			},
			expected: "executor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.Route(tt.req)
			if result != tt.expected {
				t.Errorf("Route() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNewService 测试创建Agent服务
func TestNewService(t *testing.T) {
	// 创建必要的依赖
	router := NewSimpleRouter()
	llmClient := &MockLLMClient{}
	executorAgent := NewExecutorAgent(llmClient)
	agents := []Agent{executorAgent}

	// 测试NewService函数
	svc := NewService(router, agents, nil, nil, nil, nil, llmClient)
	if svc == nil {
		t.Error("NewService() returned nil")
	}

	if svc.agents["executor"] == nil {
		t.Error("executor agent not found in service")
	}
}

// TestApplyTaskPatches 测试应用TaskPatches
func TestApplyTaskPatches(t *testing.T) {
	// 测试框架，实际测试需要数据库连接
	t.Log("TestApplyTaskPatches framework")
}
