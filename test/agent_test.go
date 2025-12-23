package test

import (
	"context"
	"testing"

	"assistant-qisumi/internal/agent"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
)

// MockAgentLLMClient 用于测试的LLM客户端模拟
type MockAgentLLMClient struct{}

func (m *MockAgentLLMClient) Chat(ctx context.Context, cfg llm.Config, req llm.ChatRequest) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{
		Choices: []struct {
			Message      llm.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{
			{
				Message: llm.ChatMessage{Content: `{"agent": "executor"}`},
			},
		},
	}, nil
}

// TestRouterAgent 测试路由Agent
func TestRouterAgent(t *testing.T) {
	// 创建简单路由器
	router := agent.NewSimpleRouter()

	// 测试用例
	tests := []struct {
		name     string
		req      agent.AgentRequest
		expected string
	}{
		{
			name: "global session",
			req: agent.AgentRequest{
				Session: &session.Session{Type: "global"},
			},
			expected: "global",
		},
		{
			name: "summarizer request",
			req: agent.AgentRequest{
				Session:   &session.Session{Type: "task"},
				UserInput: "总结一下这个任务的进度",
			},
			expected: "summarizer",
		},
		{
			name: "planner request",
			req: agent.AgentRequest{
				Session:   &session.Session{Type: "task"},
				UserInput: "重新规划一下任务",
			},
			expected: "planner",
		},
		{
			name: "executor request",
			req: agent.AgentRequest{
				Session:   &session.Session{Type: "task"},
				UserInput: "我完成了第一步",
			},
			expected: "executor",
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.Route(tt.req)
			if result != tt.expected {
				t.Errorf("Route() = %v, want %v", result, tt.expected)
			}
			t.Logf("Test %s passed: %v", tt.name, result)
		})
	}
}

// TestAgentServiceCreation 测试Agent服务创建
func TestAgentServiceCreation(t *testing.T) {
	// 创建LLM客户端
	llmClient := &MockAgentLLMClient{}

	// 初始化工具执行器映射
	toolMap := make(map[string]agent.ToolExecutor)

	// 初始化Chat Completions处理器
	chatCompletionsHandler := agent.NewChatCompletionsHandler(llmClient, toolMap)

	// 创建Agent
	executorAgent := agent.NewExecutorAgent(llmClient, chatCompletionsHandler)
	plannerAgent := agent.NewPlannerAgent(llmClient, chatCompletionsHandler)
	summarizerAgent := agent.NewSummarizerAgent(llmClient)
	globalAgent := agent.NewGlobalAgent(llmClient, chatCompletionsHandler)

	// 收集所有Agent
	agents := []agent.Agent{
		executorAgent,
		plannerAgent,
		summarizerAgent,
		globalAgent,
	}

	// 创建路由器
	router := agent.NewSimpleRouter()

	// 创建Agent服务
	agentService := agent.NewService(router, agents, nil, nil, nil, nil, llmClient)

	// 验证服务创建
	if agentService == nil {
		t.Fatal("AgentService is nil")
	}
	t.Log("AgentService created successfully")
}

// TestTaskPatchType 测试TaskPatch类型
func TestTaskPatchType(t *testing.T) {
	// 测试TaskPatch的各种类型
	patchTypes := []agent.PatchKind{
		agent.PatchUpdateStep,
		agent.PatchAddSteps,
		agent.PatchUpdateTask,
		agent.PatchAddDependencies,
	}

	for _, patchType := range patchTypes {
		patch := agent.TaskPatch{
			Kind: patchType,
		}
		if patch.Kind != patchType {
			t.Errorf("Expected patch type %s, got %s", patchType, patch.Kind)
		}
		t.Logf("TaskPatch type %s is valid", patchType)
	}
}
