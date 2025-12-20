package test

import (
	"context"
	"testing"

	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/task"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// MockTaskLLClient 用于测试的LLM客户端模拟
type MockTaskLLClient struct{}

func (m *MockTaskLLClient) Chat(ctx context.Context, cfg llm.Config, req llm.ChatRequest) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{
		Choices: []struct {
			Message      llm.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{
			{
				Message: llm.ChatMessage{
					Content: `{
						"title": "测试任务",
						"description": "这是一个测试任务",
						"priority": "medium",
						"steps": [
							{
								"title": "第一步",
								"detail": "测试第一步的详细描述",
								"estimate_minutes": 30,
								"status": "todo"
							}
						]
					}`,
				},
			},
		},
	}, nil
}

func setupTaskServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&task.Task{}, &task.TaskStep{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

// TestService_CreateFromText 测试从文本创建任务的核心逻辑
func TestService_CreateFromText(t *testing.T) {
	db := setupTaskServiceTestDB(t)
	repo := task.NewRepository(db)
	llmClient := &MockTaskLLClient{}
	service := task.NewService(repo, llmClient)

	ctx := context.Background()
	userID := uint64(1)
	rawText := "帮我安排一个测试任务，第一步是测试第一步的详细描述，预计30分钟"
	cfg := llm.Config{Model: "gpt-3.5-turbo"}

	createdTask, err := service.CreateFromText(ctx, userID, rawText, cfg)
	if err != nil {
		t.Fatalf("CreateFromText failed: %v", err)
	}

	if createdTask.Title != "测试任务" {
		t.Errorf("expected title '测试任务', got '%s'", createdTask.Title)
	}

	if len(createdTask.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(createdTask.Steps))
	}

	if createdTask.Steps[0].Title != "第一步" {
		t.Errorf("expected step title '第一步', got '%s'", createdTask.Steps[0].Title)
	}

	// 验证数据库中是否已插入
	var dbTask task.Task
	err = db.Preload("Steps").First(&dbTask, createdTask.ID).Error
	if err != nil {
		t.Fatalf("failed to find task in db: %v", err)
	}

	if dbTask.Title != "测试任务" {
		t.Errorf("db task title mismatch: expected '测试任务', got '%s'", dbTask.Title)
	}

	if len(dbTask.Steps) != 1 {
		t.Errorf("db task steps count mismatch: expected 1, got %d", len(dbTask.Steps))
	}
}

func TestService_ListTasks(t *testing.T) {
	db := setupTaskServiceTestDB(t)
	repo := task.NewRepository(db)
	service := task.NewService(repo, nil)

	ctx := context.Background()
	userID := uint64(1)

	// 插入测试数据
	tasks := []task.Task{
		{UserID: userID, Title: "任务1", Status: "todo"},
		{UserID: userID, Title: "任务2", Status: "todo"},
		{UserID: 2, Title: "其他用户的任务", Status: "todo"},
	}
	for i := range tasks {
		if err := db.Create(&tasks[i]).Error; err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	got, err := service.ListTasks(ctx, userID)
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(got))
	}
}

func TestService_GetTask(t *testing.T) {
	db := setupTaskServiceTestDB(t)
	repo := task.NewRepository(db)
	service := task.NewService(repo, nil)

	ctx := context.Background()
	userID := uint64(1)

	testTask := task.Task{
		UserID: userID,
		Title:  "详情任务",
		Steps: []task.TaskStep{
			{Title: "步骤1", OrderIndex: 1},
			{Title: "步骤2", OrderIndex: 2},
		},
	}
	if err := db.Create(&testTask).Error; err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	got, err := service.GetTask(ctx, userID, testTask.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if got.Title != "详情任务" {
		t.Errorf("expected title '详情任务', got '%s'", got.Title)
	}

	if len(got.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(got.Steps))
	}
}

func TestService_UpdateTask(t *testing.T) {
	db := setupTaskServiceTestDB(t)
	repo := task.NewRepository(db)
	service := task.NewService(repo, nil)

	ctx := context.Background()
	userID := uint64(1)

	testTask := task.Task{UserID: userID, Title: "旧标题", Status: "todo"}
	db.Create(&testTask)

	newTitle := "新标题"
	err := service.UpdateTask(ctx, userID, testTask.ID, task.UpdateTaskFields{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	var updated task.Task
	db.First(&updated, testTask.ID)
	if updated.Title != newTitle {
		t.Errorf("expected title '%s', got '%s'", newTitle, updated.Title)
	}
}

func TestService_UpdateStep(t *testing.T) {
	db := setupTaskServiceTestDB(t)
	repo := task.NewRepository(db)
	service := task.NewService(repo, nil)

	ctx := context.Background()
	userID := uint64(1)

	testTask := task.Task{
		UserID: userID,
		Title:  "任务",
		Steps: []task.TaskStep{
			{Title: "旧步骤", Status: "todo"},
		},
	}
	db.Create(&testTask)
	stepID := testTask.Steps[0].ID

	newStatus := "done"
	err := service.UpdateStep(ctx, userID, testTask.ID, stepID, task.UpdateStepFields{
		Status: &newStatus,
	})
	if err != nil {
		t.Fatalf("UpdateStep failed: %v", err)
	}

	var updatedStep task.TaskStep
	db.First(&updatedStep, stepID)
	if updatedStep.Status != newStatus {
		t.Errorf("expected status '%s', got '%s'", newStatus, updatedStep.Status)
	}
}
