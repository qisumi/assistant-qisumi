package agent

import (
	"fmt"
	"strings"
)

func buildFallbackAssistantMessage(agentName string, patches []TaskPatch) string {
	if len(patches) == 0 {
		return defaultAssistantMessage(agentName)
	}

	var parts []string
	var taskParts []string
	var hasTaskUpdate bool
	stepStatusCounts := make(map[string]int)
	stepUpdateCount := 0
	addedStepsCount := 0
	addedDependenciesCount := 0
	focusTodayCount := 0
	createdTaskCount := 0

	for _, patch := range patches {
		switch patch.Kind {
		case PatchUpdateTask:
			if patch.UpdateTask == nil {
				continue
			}
			hasTaskUpdate = true
			fields := patch.UpdateTask.Fields
			if fields.Status != nil {
				taskParts = append(taskParts, fmt.Sprintf("任务状态设为%s", labelForTaskStatus(*fields.Status)))
			}
			if fields.Priority != nil {
				taskParts = append(taskParts, fmt.Sprintf("优先级调整为%s", labelForPriority(*fields.Priority)))
			}
			if fields.DueAt != nil {
				taskParts = append(taskParts, "更新了截止时间")
			}
			if fields.Title != nil {
				taskParts = append(taskParts, "更新了任务标题")
			}
			if fields.Description != nil {
				taskParts = append(taskParts, "更新了任务描述")
			}
			if fields.IsFocusToday != nil {
				if *fields.IsFocusToday {
					taskParts = append(taskParts, "设为今日重点")
				} else {
					taskParts = append(taskParts, "取消今日重点")
				}
			}

		case PatchUpdateStep:
			if patch.UpdateStep == nil {
				continue
			}
			stepUpdateCount++
			if patch.UpdateStep.Fields.Status != nil {
				stepStatusCounts[*patch.UpdateStep.Fields.Status]++
			}

		case PatchAddSteps:
			if patch.AddSteps == nil {
				continue
			}
			addedStepsCount += len(patch.AddSteps.StepsToInsert)

		case PatchAddDependencies:
			if patch.AddDependencies == nil {
				continue
			}
			addedDependenciesCount += len(patch.AddDependencies.Items)

		case PatchMarkTasksFocusToday:
			if patch.MarkTasksFocusToday == nil {
				continue
			}
			focusTodayCount += len(patch.MarkTasksFocusToday.TaskIDs)

		case PatchCreateTask:
			createdTaskCount++
		}
	}

	if hasTaskUpdate {
		if len(taskParts) > 0 {
			parts = append(parts, strings.Join(taskParts, "，"))
		} else {
			parts = append(parts, "已更新任务信息")
		}
	}

	if stepUpdateCount > 0 {
		statusParts := buildStepStatusParts(stepStatusCounts)
		if len(statusParts) > 0 {
			parts = append(parts, "已将"+strings.Join(statusParts, "，"))
		} else {
			parts = append(parts, fmt.Sprintf("已更新 %d 个步骤", stepUpdateCount))
		}
	}

	if addedStepsCount > 0 {
		parts = append(parts, fmt.Sprintf("已新增 %d 个步骤", addedStepsCount))
	}
	if addedDependenciesCount > 0 {
		parts = append(parts, fmt.Sprintf("已新增 %d 条依赖关系", addedDependenciesCount))
	}
	if focusTodayCount > 0 {
		parts = append(parts, fmt.Sprintf("已将 %d 个任务设为今日重点", focusTodayCount))
	}
	if createdTaskCount > 0 {
		parts = append(parts, fmt.Sprintf("已创建 %d 个任务", createdTaskCount))
	}

	if len(parts) == 0 {
		return defaultAssistantMessage(agentName)
	}

	return strings.Join(parts, "；") + "。如需继续调整请告诉我。"
}

func buildStepStatusParts(statusCounts map[string]int) []string {
	if len(statusCounts) == 0 {
		return nil
	}

	// 定义顺序：按重要性从高到低
	order := []string{"done", "in_progress", "todo", "blocked", "locked"}
	parts := make([]string, 0, len(statusCounts))
	seen := make(map[string]bool)

	// 按预定义顺序添加
	for _, status := range order {
		if count := statusCounts[status]; count > 0 {
			parts = append(parts, fmt.Sprintf("%d 个步骤标记为%s", count, labelForStepStatus(status)))
			seen[status] = true
		}
	}

	// 添加未在预定义顺序中的状态
	for status, count := range statusCounts {
		if !seen[status] {
			parts = append(parts, fmt.Sprintf("%d 个步骤标记为%s", count, labelForStepStatus(status)))
		}
	}

	return parts
}

func defaultAssistantMessage(agentName string) string {
	switch agentName {
	case "planner":
		return "已收到你的规划需求。如果需要调整步骤或时间安排，请告诉我。"
	case "global":
		return "已收到你的问题。如果需要我继续整理任务安排，请告诉我。"
	case "summarizer":
		return "已整理当前任务信息。如需更详细的总结，请告诉我。"
	default:
		return "已收到。如需继续调整请告诉我。"
	}
}

func labelForTaskStatus(status string) string {
	labels := map[string]string{
		"todo":        "待办",
		"in_progress": "进行中",
		"done":        "已完成",
		"cancelled":   "已取消",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

func labelForStepStatus(status string) string {
	labels := map[string]string{
		"locked":      "已锁定",
		"todo":        "待办",
		"in_progress": "进行中",
		"done":        "已完成",
		"blocked":     "受阻",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

func labelForPriority(priority string) string {
	labels := map[string]string{
		"low":    "低",
		"medium": "中",
		"high":   "高",
	}
	if label, ok := labels[priority]; ok {
		return label
	}
	return priority
}
