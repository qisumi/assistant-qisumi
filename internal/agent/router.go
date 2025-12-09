package agent

import (
	"strings"
)

type Router interface {
	Route(req AgentRequest) string
}

type SimpleRouter struct{}

func NewSimpleRouter() *SimpleRouter {
	return &SimpleRouter{}
}

func (r *SimpleRouter) Route(req AgentRequest) string {
	text := strings.ToLower(req.UserInput)

	if req.Session.Type == "global" {
		return "global"
	}

	if strings.Contains(text, "总结") || strings.Contains(text, "overview") {
		return "summarizer"
	}
	if strings.Contains(text, "重排") || strings.Contains(text, "reschedule") || strings.Contains(text, "重新规划") {
		return "planner"
	}
	// 默认执行器
	return "executor"
}
