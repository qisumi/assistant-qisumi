package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// TaskCreationInput 任务创建的输入
type TaskCreationInput struct {
	RawText string    // 用户输入的原始文本
	Now     time.Time // 当前时间
}

// TaskCreationOutput 任务创建的输出（从 LLM 解析）
type TaskCreationOutput struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	DueAt       *FlexibleTime `json:"due_at,omitempty"`
	Priority    string        `json:"priority"`
	Steps       []StepData    `json:"steps"`
}

// StepData 用于解析 LLM 生成的步骤数据
type StepData struct {
	Title           string `json:"title"`
	Detail          string `json:"detail"`
	EstimateMinutes int    `json:"estimate_minutes"`
	OrderIndex      int    `json:"order_index"`
}

// ParseTaskCreationResponse 解析 LLM 返回的任务创建响应
func ParseTaskCreationResponse(content string) (*TaskCreationOutput, error) {
	content = ExtractJSON(content)
	var output TaskCreationOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse task creation response: %w, content: %s", err, content)
	}
	return &output, nil
}

// ToTask 将 TaskCreationOutput 转换为 Task 对象
func (o *TaskCreationOutput) ToTask(userID uint64) *Task {
	steps := make([]TaskStep, len(o.Steps))
	for i, s := range o.Steps {
		est := s.EstimateMinutes
		steps[i] = TaskStep{
			Title:       s.Title,
			Detail:      s.Detail,
			EstimateMin: &est,
			OrderIndex:  i,
			Status:      "todo",
		}
	}

	return &Task{
		UserID:      userID,
		Title:       o.Title,
		Description: o.Description,
		Status:      "todo",
		Priority:    o.Priority,
		DueAt:       o.DueAt,
		Steps:       steps,
	}
}

// ToNewStepRecords 将步骤数据转换为 NewStepRecord 切片
func (o *TaskCreationOutput) ToNewStepRecords() []NewStepRecord {
	records := make([]NewStepRecord, 0, len(o.Steps))
	for _, s := range o.Steps {
		est := s.EstimateMinutes
		records = append(records, NewStepRecord{
			Title:           s.Title,
			Detail:          s.Detail,
			EstimateMinutes: &est,
		})
	}
	return records
}

// DueAtString 返回 DueAt 的 RFC3339 字符串表示
func (o *TaskCreationOutput) DueAtString() *string {
	if o.DueAt == nil {
		return nil
	}
	s := o.DueAt.ToTime().Format(time.RFC3339)
	return &s
}
