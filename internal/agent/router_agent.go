package agent

import (
	"context"
	"encoding/json"
	"strings"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/logger"

	"go.uber.org/zap"
)

type RouterAgent struct {
	llmClient llm.Client
}

func NewRouterAgent(llmClient llm.Client) *RouterAgent {
	return &RouterAgent{llmClient: llmClient}
}

func (a *RouterAgent) Name() string {
	return "router"
}

func (a *RouterAgent) Handle(req AgentRequest) (*AgentResponse, error) {
	// 1. 首先尝试 rule-based 路由
	selectedAgent := a.ruleBasedRoute(req)
	logger.Logger.Debug("Rule-based路由结果",
		zap.String("agent", selectedAgent),
		zap.String("user_input", req.UserInput),
	)

	// 2. 如果 rule-based 无法确定，或者需要更智能的判断，使用 LLM fallback
	if selectedAgent == "" {
		logger.Logger.Debug("Rule-based无法确定，使用LLM路由")
		var err error
		selectedAgent, err = a.llmBasedRoute(req)
		if err != nil {
			// 如果 LLM 路由失败，使用默认值
			logger.Logger.Warn("LLM路由失败，使用默认executor",
				zap.String("error", err.Error()),
			)
			selectedAgent = "executor"
		}
		logger.Logger.Debug("LLM路由结果",
			zap.String("agent", selectedAgent),
		)
	}

	// RouterAgent 不需要生成 assistant message 或 task patches，
	// 它的作用是在 AgentService 中选择合适的 Agent 进行处理
	// 因此这里返回空的响应，实际的路由逻辑在 AgentService 中执行
	return &AgentResponse{
		AssistantMessage: "",
		TaskPatches:      nil,
	}, nil
}

// ruleBasedRoute 根据规则路由到合适的 Agent
func (a *RouterAgent) ruleBasedRoute(req AgentRequest) string {
	text := strings.ToLower(req.UserInput)

	// 全局会话直接路由到 GlobalAgent
	if req.Session != nil && req.Session.Type == "global" {
		logger.Logger.Debug("全局会话路由到global agent")
		return "global"
	}

	// 检查关键字，确定 Agent 类型
	if strings.Contains(text, "总结") || strings.Contains(text, "overview") || strings.Contains(text, "回顾") || strings.Contains(text, "progress") {
		logger.Logger.Debug("匹配到总结关键字，路由到summarizer agent")
		return "summarizer"
	}

	if strings.Contains(text, "重新规划") || strings.Contains(text, "重排") || strings.Contains(text, "reschedule") || strings.Contains(text, "重排日程") || strings.Contains(text, "拆解") {
		logger.Logger.Debug("匹配到规划关键字，路由到planner agent")
		return "planner"
	}

	// 其他情况默认使用 executor
	logger.Logger.Debug("无匹配关键字，默认路由到executor agent")
	return "executor"
}

// llmBasedRoute 使用 LLM 进行智能路由
func (a *RouterAgent) llmBasedRoute(req AgentRequest) (string, error) {
	logger.Logger.Info("开始LLM路由",
		zap.String("user_input", req.UserInput),
	)

	// 构造会话类型信息
	sessionType := "task"
	if req.Session != nil {
		sessionType = req.Session.Type
	}

	// 构造是否绑定任务信息
	hasTask := "true"
	if req.Task == nil {
		hasTask = "false"
	}

	// 构造 messages
	messages := []llm.Message{
		{
			Role: "system",
			Content: `你是一个路由助手（Router Agent）。

你的唯一任务是：根据用户的最新输入和当前会话类型，为系统选择应该调用哪个子 Agent。

子 Agent 类型包括：
- "executor"  : 执行/进度更新类操作（标记步骤完成、修改截止时间等）
- "planner"   : 规划/重排类操作（拆解任务、重排步骤、重排日程、设置依赖等）
- "summarizer": 单任务总结类操作（进度概览、总结近期变更）
- "global"    : 跨任务规划/总结（例如「我今天要做什么」、「这周安排如何」）

输入信息：
- 会话类型：` + sessionType + `
- 是否绑定了具体任务：` + hasTask + `
- 用户最新输入的自然语言内容：` + req.UserInput + `

你的输出必须是一个 JSON 对象，格式为：
{
  "agent": "executor" | "planner" | "summarizer" | "global"
}

要求：
- 不要输出多余字段，不要输出自然语言解释。
- 在 task 会话中：
  - 如果用户问「任务进度如何」「帮我总结这个任务」，选 summarizer。
  - 如果用户说「重新规划一下、重排日程、把后面几步拆细」，选 planner。
  - 其它绝大多数更新任务进度/状态的请求，选 executor。
- 在 global 会话中：通常直接选 global，除非用户明显在说别的。`,
		},
	}

	// 构造 Chat 请求
	chatReq := llm.ChatRequest{
		Model:      req.LLMConfig.Model,
		Messages:   messages,
		ToolChoice: "none", // Router 不需要工具调用
	}

	// 调用 LLM
	logger.Logger.Debug("发送LLM路由请求",
		zap.String("model", req.LLMConfig.Model),
		zap.String("session_type", sessionType),
		zap.String("has_task", hasTask),
	)
	resp, err := a.llmClient.Chat(context.Background(), req.LLMConfig, chatReq)
	if err != nil {
		logger.Logger.Error("LLM路由请求失败",
			zap.String("error", err.Error()),
		)
		return "", err
	}
	logger.Logger.Debug("收到LLM路由响应",
		zap.Int("choices_count", len(resp.Choices)),
	)

	// 解析 LLM 响应
	if len(resp.Choices) > 0 {
		assistantMessage := resp.Choices[0].Message.Content
		logger.Logger.Debug("LLM路由响应内容",
			zap.String("content", assistantMessage),
		)

		// 解析 JSON 响应
		var routerResp struct {
			Agent string `json:"agent"`
		}

		if err := json.Unmarshal([]byte(assistantMessage), &routerResp); err == nil {
			// 验证返回的 Agent 类型是否有效
			validAgents := map[string]bool{
				"executor":   true,
				"planner":    true,
				"summarizer": true,
				"global":     true,
			}

			if validAgents[routerResp.Agent] {
				logger.Logger.Info("LLM路由成功",
					zap.String("agent", routerResp.Agent),
				)
				return routerResp.Agent, nil
			} else {
				logger.Logger.Warn("LLM返回无效的agent类型",
					zap.String("agent", routerResp.Agent),
				)
			}
		} else {
			logger.Logger.Warn("解析LLM路由响应失败",
				zap.String("error", err.Error()),
				zap.String("content", assistantMessage),
			)
		}
	}

	// 默认返回 executor
	logger.Logger.Warn("LLM路由无效，使用默认executor")
	return "executor", nil
}
