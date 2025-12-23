package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"
)

// historyToLLMMessages 将内部的对话消息转换为 LLM 消息（简单版：只保留 role+content）
func historyToLLMMessages(history []session.Message) []llm.Message {
	msgs := make([]llm.Message, 0, len(history))
	for _, m := range history {
		role := m.Role
		// 确保只出现 user/assistant/system 三种
		if role != "user" && role != "assistant" && role != "system" {
			continue
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		msgs = append(msgs, llm.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return msgs
}

// BuildExecutorMessages 构造 ExecutorAgent 的 messages：
// - system: ExecutorSystemPrompt
// - system: 当前任务 JSON
// - system: 当前时间
// - 历史消息（可选）
// - user: 最新输入
func BuildExecutorMessages(t *task.Task, history []session.Message, userInput string, now time.Time) ([]llm.Message, error) {
	taskJSON, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	msgs := []llm.Message{
		{
			Role:    "system",
			Content: ExecutorSystemPrompt,
		},
		{
			Role: "system",
			Content: fmt.Sprintf(
				"当前任务状态（JSON，只读）：\n%s",
				taskJSON,
			),
		},
		{
			Role:    "system",
			Content: fmt.Sprintf("当前时间 now: %s", now.Format(time.RFC3339)),
		},
	}

	// 历史消息
	msgs = append(msgs, historyToLLMMessages(history)...)

	// 最新用户消息
	msgs = append(msgs, llm.Message{
		Role:    "user",
		Content: userInput,
	})

	return msgs, nil
}

// BuildPlannerMessages 构造 PlannerAgent 的 messages。
// 这里同样包含：系统提示词 + 当前任务 JSON + 当前时间 + 历史 + 最新 user。
func BuildPlannerMessages(t *task.Task, history []session.Message, userInput string, now time.Time) ([]llm.Message, error) {
	taskJSON, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	msgs := []llm.Message{
		{
			Role:    "system",
			Content: PlannerSystemPrompt,
		},
		{
			Role: "system",
			Content: fmt.Sprintf(
				"当前任务结构（JSON，只读）：\n%s",
				taskJSON,
			),
		},
		{
			Role:    "system",
			Content: fmt.Sprintf("当前时间 now: %s", now.Format(time.RFC3339)),
		},
	}

	msgs = append(msgs, historyToLLMMessages(history)...)

	msgs = append(msgs, llm.Message{
		Role:    "user",
		Content: userInput,
	})

	return msgs, nil
}
