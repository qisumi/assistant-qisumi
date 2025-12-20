package agent

import (
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/task"
	"encoding/json"
	"fmt"
)

// 从一次 ChatResponse 中，提取所有 tool_calls 并转为 TaskPatch 列表
func BuildPatchesFromToolCalls(resp *llm.ChatResponse) ([]TaskPatch, error) {
	var patches []TaskPatch

	if len(resp.Choices) == 0 {
		return nil, nil
	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) == 0 {
		return nil, nil
	}

	for _, tc := range msg.ToolCalls {
		if tc.Type != "function" {
			continue
		}
		name := tc.Function.Name
		argsJSON := tc.Function.Arguments

		switch name {
		case "update_task":
			var args UpdateTaskArgs
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return nil, fmt.Errorf("update_task args decode: %w", err)
			}
			patches = append(patches, TaskPatch{
				Kind: PatchUpdateTask,
				UpdateTask: &UpdateTaskPatch{
					TaskID: args.TaskID,
					Fields: args.Fields,
				},
			})

		case "update_steps":
			var args UpdateStepsArgs
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return nil, fmt.Errorf("update_steps args decode: %w", err)
			}
			for _, u := range args.Updates {
				patches = append(patches, TaskPatch{
					Kind: PatchUpdateStep,
					UpdateStep: &UpdateStepPatch{
						TaskID: args.TaskID,
						StepID: u.StepID,
						Fields: u.Fields,
					},
				})
			}

		case "add_steps":
			var args AddStepsArgs
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return nil, fmt.Errorf("add_steps args decode: %w", err)
			}
			records := make([]task.NewStepRecord, 0, len(args.Steps))
			for _, s := range args.Steps {
				records = append(records, task.NewStepRecord{
					Title:             s.Title,
					Detail:            s.Detail,
					EstimateMinutes:   s.EstimateMinutes,
					InsertAfterStepID: s.InsertAfterStepID,
				})
			}
			patches = append(patches, TaskPatch{
				Kind: PatchAddSteps,
				AddSteps: &AddStepsPatch{
					TaskID:        args.TaskID,
					ParentStepID:  args.ParentStepID,
					StepsToInsert: records,
				},
			})

		case "add_dependencies":
			var args AddDependenciesArgs
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return nil, fmt.Errorf("add_dependencies args decode: %w", err)
			}
			items := make([]task.DependencyItem, 0, len(args.Items))
			for _, it := range args.Items {
				items = append(items, task.DependencyItem{
					PredecessorTaskID: it.PredecessorTaskID,
					PredecessorStepID: it.PredecessorStepID,
					SuccessorTaskID:   it.SuccessorTaskID,
					SuccessorStepID:   it.SuccessorStepID,
					Condition:         it.Condition,
					Action:            it.Action,
				})
			}
			patches = append(patches, TaskPatch{
				Kind: PatchAddDependencies,
				AddDependencies: &AddDependenciesPatch{
					Items: items,
				},
			})

		case "mark_tasks_focus_today":
			var args MarkTasksFocusTodayArgs
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return nil, fmt.Errorf("mark_tasks_focus_today args decode: %w", err)
			}
			patches = append(patches, TaskPatch{
				Kind: PatchMarkTasksFocusToday,
				MarkTasksFocusToday: &MarkTasksFocusTodayPatch{
					TaskIDs: args.TaskIDs,
				},
			})

		default:
			// 未知 function，可以忽略或记录日志
			// log.Printf("unknown tool function: %s", name)
		}
	}

	return patches, nil
}
