package agent

import (
	"context"
	"time"

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
	// 1. 构造 messages + tools 调用 LLM
	messages := []llm.Message{
		{
			Role: "system",
			Content: `你是一个跨任务日程规划助手（Global Agent）。

你收到的是：
- 用户今天/本周的任务概览（由系统消息提供，包含多个 task 列表）
  - 每个任务包含：title, status, priority, due_at
  - 每个任务下有若干关键步骤（title, status, estimate_minutes, planned_start/planned_end）
- 用户的提问，例如：
  - 「我今天要做什么？」
  - 「帮我看一下这周的安排」
  - 「有没有已经过期但没完成的任务？」

你的职责：
1. 基于给出的任务数据，为用户生成一个清晰的「计划说明」，例如：
   - 今日待办清单（按优先级和紧迫度排序）
   - 已过期但未完成的任务提醒
   - 建议的执行顺序（可以按时间或能量水平来安排）
2. 你可以使用工具：
   - mark_tasks_focus_today：标记今天重点关注的任务（如果用户有此意图）
   - update_task / update_steps：仅在用户明确要求修改时使用（例如「帮我把某任务优先级调高」）
3. 输出中尽量包含结构化层次：
   - 第一部分：今日重点任务
   - 第二部分：可选任务/轻量任务
   - 第三部分：过期任务的提醒

注意：
- 你不会自己查询数据库，你只能使用系统给你的任务数据。
- 不要随意修改任务状态，除非用户有明确指令。`,
		},
	}

	// 添加当前时间信息
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: "当前时间 now: " + req.Now.Format(time.RFC3339),
	})

	// 添加历史消息
	messages = append(messages, historyToLLMMessages(req.Messages)...)

	// 添加当前用户输入
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: req.UserInput,
	})

	// 定义可用工具
	tools := llm.GlobalTools()

	// 构造Chat请求
	chatReq := llm.ChatRequest{
		Model:      req.LLMConfig.Model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
	}

	// 调用LLM
	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		return nil, err
	}

	// 2. 处理 LLM 响应，生成 assistant 回复文本和 TaskPatches
	var assistantMessage string
	if len(resp.Choices) > 0 {
		assistantMessage = resp.Choices[0].Message.Content
	}

	taskPatches, err := BuildPatchesFromToolCalls(resp)
	if err != nil {
		// 记录错误但继续返回 assistant 消息
	}

	return &AgentResponse{
		AssistantMessage: assistantMessage,
		TaskPatches:      taskPatches,
	}, nil
}
