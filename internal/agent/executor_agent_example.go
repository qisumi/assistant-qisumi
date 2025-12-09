package agent

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"
)

// ExecutorAgentExample demonstrates the complete flow of using ExecutorAgent
// including message construction, LLM call, tool response parsing, and task patch application
func ExecutorAgentExample() {
	// This is a simplified example demonstrating the ExecutorAgent workflow
	// In production, this would be integrated with the actual service layer

	// 1. Mock user input and context
	userInput := "I've completed the '整理相关工作' step."

	// 2. Mock task data (would come from database in production)
	// Create pointer variables for estimate minutes
	estimate120 := 120
	estimate300 := 300
	mockDueAt := time.Now().AddDate(0, 0, 7) // 7 days from now

	taskData := &task.Task{
		ID:          1,
		UserID:      123,
		Title:       "写 AIGC 论文",
		Description: "完成一篇关于 AIGC 的研究论文",
		Status:      "in_progress",
		Priority:    "high",
		DueAt:       &mockDueAt, // Mock due date
		Steps: []task.Step{
			{
				ID:          12,
				TaskID:      1,
				OrderIndex:  0,
				Title:       "整理相关工作",
				Detail:      "收集和整理相关研究文献",
				Status:      "in_progress",
				EstimateMin: &estimate120,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          13,
				TaskID:      1,
				OrderIndex:  1,
				Title:       "撰写论文初稿",
				Detail:      "根据整理的文献撰写论文初稿",
				Status:      "todo",
				EstimateMin: &estimate300,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	// 3. Mock LLM client (would be actual client in production)
	mockLLMClient := &mockLLMClient{
		// This mock client simulates the LLM response for our example
		response: llm.ChatResponse{
			Choices: []struct {
				Message      llm.ChatMessage `json:"message"`
				FinishReason string          `json:"finish_reason"`
			}{
				{
					Message: llm.ChatMessage{
						Role:    "assistant",
						Content: "",
						ToolCalls: []llm.ToolCall{
							{
								ID:   "call_update_steps_1",
								Type: "function",
								Function: llm.ToolCallFunc{
									Name:      "update_steps",
									Arguments: `{"task_id": 1, "updates": [{"step_id": 12, "fields": {"status": "done"}}]}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		},
	}

	// 4. Create and configure the ExecutorAgent
	executorAgent := NewExecutorAgent(mockLLMClient)

	// 5. Construct the AgentRequest
	req := AgentRequest{
		UserID:    123,
		Session:   nil, // Would be populated in production
		Task:      taskData,
		Messages:  []session.Message{}, // Would include history in production
		UserInput: userInput,
		Now:       time.Now(),
		LLMConfig: llm.Config{
			BaseURL: "https://api.openai.com/v1",
			Model:   "gpt-4",
			APIKey:  "sk-xxx", // Would be encrypted in production
		},
	}

	// 6. Call the Agent's Handle method
	resp, err := executorAgent.Handle(req)
	if err != nil {
		log.Fatalf("Error handling request: %v", err)
	}

	// 7. Process the result
	log.Printf("Assistant Message: %s", resp.AssistantMessage)
	log.Printf("Generated %d Task Patches", len(resp.TaskPatches))

	for i, patch := range resp.TaskPatches {
		log.Printf("Patch %d: Type=%s", i+1, patch.Type)
		patchJSON, _ := json.Marshal(patch.Payload)
		log.Printf("Patch %d: Payload=%s", i+1, string(patchJSON))

		// 8. Apply the patch to the task data (in production, this would update the database)
		applyPatchToTask(taskData, patch)
	}

	// 9. Log the updated task state
	log.Printf("Updated Task Status: %s", taskData.Status)
	for _, step := range taskData.Steps {
		log.Printf("Step %d: %s - Status: %s", step.ID, step.Title, step.Status)
	}
}

// applyPatchToTask applies a task patch to the task data (in-memory for this example)
func applyPatchToTask(task *task.Task, patch TaskPatch) {
	switch patch.Type {
	case "update_task":
		// Update task fields
		if fields, ok := patch.Payload["fields"].(map[string]interface{}); ok {
			if status, ok := fields["status"].(string); ok {
				task.Status = status
			}
			if title, ok := fields["title"].(string); ok {
				task.Title = title
			}
			// ... other task fields
		}

	case "update_step":
		// Update step fields
		if stepID, ok := patch.Payload["step_id"].(uint64); ok {
			if fields, ok := patch.Payload["fields"].(map[string]interface{}); ok {
				for i, step := range task.Steps {
					if step.ID == stepID {
						if status, ok := fields["status"].(string); ok {
							task.Steps[i].Status = status
						}
						if title, ok := fields["title"].(string); ok {
							task.Steps[i].Title = title
						}
						// ... other step fields
						break
					}
				}
			}
		}

		// ... other patch types
	}
}

// mockLLMClient is a mock implementation of llm.Client for testing purposes
type mockLLMClient struct {
	response llm.ChatResponse
}

func (m *mockLLMClient) Chat(ctx context.Context, config llm.Config, req llm.ChatRequest) (*llm.ChatResponse, error) {
	// Log the request for demonstration purposes
	log.Printf("Mock LLM Client received request with %d messages", len(req.Messages))
	log.Printf("Available tools: %d", len(req.Tools))

	// Return the predefined mock response
	return &m.response, nil
}

// This function can be called from main.go or a test file to demonstrate the workflow
func RunExecutorAgentExample() {
	log.Println("=== ExecutorAgent Example ===")
	ExecutorAgentExample()
	log.Println("=== Example Completed ===")
}
