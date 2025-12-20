package test

import (
	"context"
	"testing"
	"time"

	"assistant-qisumi/internal/db"
	"assistant-qisumi/internal/llm"
	"assistant-qisumi/internal/task"
)

// TestFullTaskWorkflow 测试完整的任务工作流
func TestFullTaskWorkflow(t *testing.T) {
	// 获取测试配置
	cfg, err := GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// 初始化数据库连接
	dbConn, err := db.InitDB(cfg.DB)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	gormDB, err := db.InitGORM(dbConn, cfg.DB.Type)
	if err != nil {
		t.Fatalf("Failed to init GORM: %v", err)
	}

	// 创建测试用户
	if err := SetupTestUser(dbConn); err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	// 创建任务仓库
	taskRepo := task.NewRepository(gormDB)

	// 创建LLM客户端
	llmClient := llm.NewHTTPClient()

	// 创建任务服务 - 目前在集成测试中不需要直接使用
	_ = task.NewService(taskRepo, llmClient)

	// 测试用户ID
	userID := TestUserInfo.ID

	// 阶段1：创建任务和步骤
	t.Log("=== Stage 1: Creating Task and Steps ===")

	// 创建测试任务
	dueAt := time.Now().Add(7 * 24 * time.Hour)
	testTask := task.Task{
		UserID:      userID,
		Title:       "Integration Test Task",
		Description: "This is a test task for integration testing",
		Status:      "todo",
		Priority:    "high",
		DueAt:       &dueAt,
		Steps: []task.TaskStep{
			{
				TaskID:     0, // 会被自动设置
				OrderIndex: 0,
				Title:      "Step 1",
				Detail:     "First step of the task",
				Status:     "todo",
			},
			{
				TaskID:     0, // 会被自动设置
				OrderIndex: 1,
				Title:      "Step 2",
				Detail:     "Second step of the task",
				Status:     "todo",
			},
			{
				TaskID:     0, // 会被自动设置
				OrderIndex: 2,
				Title:      "Step 3",
				Detail:     "Third step of the task",
				Status:     "locked", // 初始锁定状态
			},
		},
	}

	// 插入任务和步骤
	if err := taskRepo.InsertTaskWithSteps(context.Background(), &testTask); err != nil {
		t.Fatalf("Failed to insert task with steps: %v", err)
	}
	t.Logf("Created task: %+v", testTask)

	// 阶段2：获取任务和步骤
	t.Log("=== Stage 2: Retrieving Task and Steps ===")

	// 获取任务和步骤
	retrievedTask, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get task with steps: %v", err)
	}
	t.Logf("Retrieved task: %+v", retrievedTask)
	if len(retrievedTask.Steps) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(retrievedTask.Steps))
	}

	// 阶段3：更新步骤状态
	t.Log("=== Stage 3: Updating Step Status ===")

	// 更新步骤1的状态为done
	statusDone := "done"
	if err := taskRepo.ApplyUpdateStepFields(context.Background(), userID, testTask.ID, retrievedTask.Steps[0].ID, task.UpdateStepFields{
		Status: &statusDone,
	}); err != nil {
		t.Fatalf("Failed to update step status: %v", err)
	}

	// 验证步骤状态更新
	updatedTask, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if updatedTask.Steps[0].Status != "done" {
		t.Fatalf("Expected step 0 to be 'done', got '%s'", updatedTask.Steps[0].Status)
	}
	t.Logf("Step 0 status updated to 'done' successfully")

	// 阶段4：更新任务状态
	t.Log("=== Stage 4: Updating Task Status ===")

	// 更新任务状态为in_progress
	statusInProgress := "in_progress"
	if err := taskRepo.ApplyUpdateTaskFields(context.Background(), userID, testTask.ID, task.UpdateTaskFields{
		Status: &statusInProgress,
	}); err != nil {
		t.Fatalf("Failed to update task status: %v", err)
	}

	// 验证任务状态更新
	updatedTask2, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if updatedTask2.Status != "in_progress" {
		t.Fatalf("Expected task to be 'in_progress', got '%s'", updatedTask2.Status)
	}
	t.Logf("Task status updated to 'in_progress' successfully")

	// 阶段5：添加新步骤
	t.Log("=== Stage 5: Adding New Step ===")

	// 添加新步骤
	newStep := task.TaskStep{
		TaskID:     testTask.ID,
		OrderIndex: 3,
		Title:      "Step 4",
		Detail:     "Fourth step of the task",
		Status:     "todo",
	}
	if err := taskRepo.AddStep(context.Background(), &newStep); err != nil {
		t.Fatalf("Failed to add new step: %v", err)
	}
	t.Logf("Added new step: %+v", newStep)

	// 验证新步骤已添加
	updatedTask3, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if len(updatedTask3.Steps) != 4 {
		t.Fatalf("Expected 4 steps, got %d", len(updatedTask3.Steps))
	}
	t.Logf("Total steps after adding new step: %d", len(updatedTask3.Steps))

	// 阶段6：测试任务列表
	t.Log("=== Stage 6: Testing Task List ===")

	// 获取用户的任务列表
	tasks, err := taskRepo.ListTasks(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	if len(tasks) < 1 {
		t.Fatalf("Expected at least 1 task, got %d", len(tasks))
	}
	t.Logf("Found %d tasks for user %d", len(tasks), userID)

	t.Log("=== Integration Test Completed Successfully ===")
}
