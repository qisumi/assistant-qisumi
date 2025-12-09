package tool

import (
	"encoding/json"
)

// ToolFunction 工具函数定义
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolCall 工具调用请求
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolCallFunc `json:"function"`
}

// ToolCallFunc 工具调用函数
type ToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolResponse 工具调用响应
type ToolResponse struct {
	ToolCallID string      `json:"tool_call_id"`
	Name       string      `json:"name"`
	Content    interface{} `json:"content"`
}

// ToolArguments 工具参数解析接口
type ToolArguments interface {
	Validate() error
}

// ParseToolArguments 解析工具调用参数
func ParseToolArguments(arguments string, dest ToolArguments) error {
	return json.Unmarshal([]byte(arguments), dest)
}
