package agent

import (
	"encoding/json"
)

// 工具执行器实现

// NoOpExecutor 不执行实际操作的工具执行器
// 因为实际的数据库更新是通过 TaskPatch 完成的，所以这里只需要返回成功响应
type NoOpExecutor struct{}

func (e *NoOpExecutor) Execute(args string) (interface{}, error) {
	// 解析参数以验证格式
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return nil, err
	}

	// 返回成功响应
	return map[string]interface{}{
		"success": true,
	}, nil
}

// NewToolExecutors 创建所有工具执行器的映射
func NewToolExecutors() map[string]ToolExecutor {
	return map[string]ToolExecutor{
		"update_task":            &NoOpExecutor{},
		"update_steps":           &NoOpExecutor{},
		"add_steps":              &NoOpExecutor{},
		"add_dependencies":       &NoOpExecutor{},
		"mark_tasks_focus_today": &NoOpExecutor{},
	}
}
