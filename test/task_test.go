package test

import (
	"context"
	"testing"
	"time"

	"assistant-qisumi/internal/db"
	"assistant-qisumi/internal/task"
)

// TestTaskRepository 测试任务仓库的基本功能
func TestTaskRepository(t *testing.T) {
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

	// 初始化 GORM
	gormDB, err := db.InitGORM(dbConn, cfg.DB.Type)
	if err != nil {
		t.Fatalf("Failed to init GORM: %v", err)
	}

	// 执行自动迁移
	if err := db.AutoMigrate(gormDB); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	// 创建测试用户
	if err := SetupTestUser(dbConn); err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	// 创建任务仓库
	taskRepo := task.NewRepository(gormDB)

	// 测试用户ID
	userID := TestUserInfo.ID

	// 测试1：创建任务和步骤
	t.Log("Test 1: Creating Task and Steps")

	// 创建测试任务
	dueAt := task.FlexibleTime{Time: time.Now().Add(24 * time.Hour)}
	testTask := task.Task{
		UserID:      userID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "todo",
		Priority:    "medium",
		DueAt:       &dueAt,
		Steps: []task.TaskStep{
			{
				TaskID:     0, // 会被自动设置
				OrderIndex: 0,
				Title:      "Test Step 1",
				Detail:     "First step",
				Status:     "todo",
			},
			{
				TaskID:     0, // 会被自动设置
				OrderIndex: 1,
				Title:      "Test Step 2",
				Detail:     "Second step",
				Status:     "todo",
			},
		},
	}

	// 插入任务和步骤
	if err := taskRepo.InsertTaskWithSteps(context.Background(), &testTask); err != nil {
		t.Fatalf("Failed to insert task with steps: %v", err)
	}
	t.Logf("Created task with ID: %d", testTask.ID)
	if len(testTask.Steps) != 2 {
		t.Fatalf("Expected 2 steps, got %d", len(testTask.Steps))
	}

	// 测试2：获取任务和步骤
	t.Log("Test 2: Getting Task and Steps")

	// 获取任务和步骤
	retrievedTask, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get task with steps: %v", err)
	}
	t.Logf("Retrieved task: %+v", retrievedTask)
	if retrievedTask.ID != testTask.ID {
		t.Fatalf("Task ID mismatch: expected %d, got %d", testTask.ID, retrievedTask.ID)
	}
	if len(retrievedTask.Steps) != 2 {
		t.Fatalf("Expected 2 steps, got %d", len(retrievedTask.Steps))
	}

	// 测试3：标记今日重点
	t.Log("Test 3: Marking Tasks Focus Today")
	if err := taskRepo.MarkTasksFocusToday(context.Background(), userID, []uint64{testTask.ID}); err != nil {
		t.Fatalf("Failed to mark task focus today: %v", err)
	}

	// 验证标记结果
	retrievedTask, err = taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get task after marking focus today: %v", err)
	}
	if !retrievedTask.IsFocusToday {
		t.Fatalf("Expected task to be focus today, but it's not")
	}

	// 测试4：更新任务
	t.Log("Test 4: Updating Task")

	// 更新任务标题和取消今日重点
	newTitle := "Updated Test Task"
	isFocusFalse := false
	if err := taskRepo.ApplyUpdateTaskFields(context.Background(), userID, testTask.ID, task.UpdateTaskFields{
		Title:        &newTitle,
		IsFocusToday: &isFocusFalse,
	}); err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// 验证任务更新
	updatedTask, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if updatedTask.Title != newTitle {
		t.Fatalf("Task title mismatch: expected %s, got %s", newTitle, updatedTask.Title)
	}
	if updatedTask.IsFocusToday {
		t.Fatalf("Expected task to NOT be focus today, but it is")
	}
	t.Logf("Task updated successfully: new title is %s", updatedTask.Title)

	// 测试5：更新步骤
	t.Log("Test 5: Updating Step")

	// 更新步骤1的状态
	statusDone := "done"
	if err := taskRepo.ApplyUpdateStepFields(context.Background(), userID, testTask.ID, retrievedTask.Steps[0].ID, task.UpdateStepFields{
		Status: &statusDone,
	}); err != nil {
		t.Fatalf("Failed to update step: %v", err)
	}

	// 验证步骤更新
	updatedTask2, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if updatedTask2.Steps[0].Status != "done" {
		t.Fatalf("Step status mismatch: expected 'done', got '%s'", updatedTask2.Steps[0].Status)
	}
	t.Logf("Step updated successfully: step 0 status is %s", updatedTask2.Steps[0].Status)

	// 测试6：添加新步骤
	t.Log("Test 6: Adding New Step")

	// 添加新步骤
	newStep := task.TaskStep{
		TaskID:     testTask.ID,
		OrderIndex: 2,
		Title:      "Test Step 3",
		Detail:     "Third step",
		Status:     "todo",
	}
	if err := taskRepo.AddStep(context.Background(), &newStep); err != nil {
		t.Fatalf("Failed to add new step: %v", err)
	}

	// 验证新步骤已添加
	updatedTask3, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if len(updatedTask3.Steps) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(updatedTask3.Steps))
	}
	t.Logf("New step added successfully: total steps now %d", len(updatedTask3.Steps))

	// 测试7：获取任务列表
	t.Log("Test 7: Getting Task List")

	// 获取任务列表
	tasks, err := taskRepo.ListTasks(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatalf("Expected at least 1 task, got %d", len(tasks))
	}
	t.Logf("Task list retrieved successfully: %d tasks found", len(tasks))

	t.Log("All tests passed!")
}

// TestTaskStepsOperations 测试任务步骤的各种操作
func TestTaskStepsOperations(t *testing.T) {
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

	// 创建测试用户
	if err := SetupTestUser(dbConn); err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	// 创建任务仓库
	gormDB, _ := db.InitGORM(dbConn, cfg.DB.Type)
	taskRepo := task.NewRepository(gormDB)

	// 测试用户ID
	userID := TestUserInfo.ID

	// 创建测试任务
	testTask := task.Task{
		UserID:      userID,
		Title:       "Test Steps Task",
		Description: "Task for testing steps operations",
		Status:      "todo",
		Priority:    "medium",
		Steps: []task.TaskStep{
			{
				TaskID:     0,
				OrderIndex: 0,
				Title:      "Initial Step",
				Detail:     "Initial step",
				Status:     "todo",
			},
		},
	}

	// 插入任务和步骤
	if err := taskRepo.InsertTaskWithSteps(context.Background(), &testTask); err != nil {
		t.Fatalf("Failed to insert task with steps: %v", err)
	}

	// 测试1：更新步骤顺序
	t.Log("Test 1: Updating Step Order")

	// 添加第二个步骤
	step2 := task.TaskStep{
		TaskID:     testTask.ID,
		OrderIndex: 1,
		Title:      "Step 2",
		Detail:     "Second step",
		Status:     "todo",
	}
	if err := taskRepo.AddStep(context.Background(), &step2); err != nil {
		t.Fatalf("Failed to add step 2: %v", err)
	}

	// 添加第三个步骤
	step3 := task.TaskStep{
		TaskID:     testTask.ID,
		OrderIndex: 2,
		Title:      "Step 3",
		Detail:     "Third step",
		Status:     "todo",
	}
	if err := taskRepo.AddStep(context.Background(), &step3); err != nil {
		t.Fatalf("Failed to add step 3: %v", err)
	}

	// 交换步骤1和步骤3的顺序
	order0 := 0
	if err := taskRepo.ApplyUpdateStepFields(context.Background(), userID, testTask.ID, step3.ID, task.UpdateStepFields{
		OrderIndex: &order0,
	}); err != nil {
		t.Fatalf("Failed to update step 3 order: %v", err)
	}

	order2 := 2
	if err := taskRepo.ApplyUpdateStepFields(context.Background(), userID, testTask.ID, testTask.Steps[0].ID, task.UpdateStepFields{
		OrderIndex: &order2,
	}); err != nil {
		t.Fatalf("Failed to update step 1 order: %v", err)
	}

	// 验证顺序更新
	updatedTask, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if len(updatedTask.Steps) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(updatedTask.Steps))
	}
	t.Logf("Steps order updated successfully")

	// 测试2：更新步骤详细信息
	t.Log("Test 2: Updating Step Details")

	// 更新步骤详细信息
	newDetail := "Updated step detail with more information"
	newEstimate := 60
	if err := taskRepo.ApplyUpdateStepFields(context.Background(), userID, testTask.ID, updatedTask.Steps[0].ID, task.UpdateStepFields{
		Detail:      &newDetail,
		EstimateMin: &newEstimate,
	}); err != nil {
		t.Fatalf("Failed to update step details: %v", err)
	}

	// 验证详细信息更新
	updatedTask2, err := taskRepo.GetTaskWithSteps(context.Background(), userID, testTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}
	if updatedTask2.Steps[0].Detail != newDetail {
		t.Fatalf("Step detail mismatch: expected %s, got %s", newDetail, updatedTask2.Steps[0].Detail)
	}
	if *updatedTask2.Steps[0].EstimateMin != newEstimate {
		t.Fatalf("Estimate mismatch: expected %d, got %d", newEstimate, *updatedTask2.Steps[0].EstimateMin)
	}
	t.Logf("Step details updated successfully")

	t.Log("All step operations tests passed!")
}
