package agent

import (
	"time"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"
)

type AgentRequest struct {
	UserID    uint64
	Session   *session.Session
	Task      *task.Task
	Messages  []session.Message
	UserInput string
	Now       time.Time
	LLMConfig llm.Config
}

type AgentResponse struct {
	AssistantMessage string      `json:"assistant_message"`
	TaskPatches      []TaskPatch `json:"task_patches"`
}

type Agent interface {
	Name() string
	Handle(req AgentRequest) (*AgentResponse, error)
}
