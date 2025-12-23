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
// - system: 依赖关系 JSON（用于判断隐含前置条件）
// - system: 当前时间
// - 历史消息（可选）
// - user: 最新输入
func BuildExecutorMessages(t *task.Task, dependencies []task.TaskDependency, history []session.Message, userInput string, now time.Time) ([]llm.Message, error) {
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
	}

	// 添加依赖关系信息
	if len(dependencies) > 0 {
		depsJSON, err := json.Marshal(dependencies)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, llm.Message{
			Role: "system",
			Content: fmt.Sprintf(
				"依赖关系（JSON，只读）：\n%s\n说明：当某个任务/步骤完成时，你需要检查这些依赖关系，自动解锁满足条件的后续任务/步骤。",
				depsJSON,
			),
		})
	}

	msgs = append(msgs, llm.Message{
		Role:    "system",
		Content: fmt.Sprintf("当前时间 now: %s", now.Format(time.RFC3339)),
	})

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
