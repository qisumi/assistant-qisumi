package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/db"
	internalHTTP "assistant-qisumi/internal/http"
	"assistant-qisumi/internal/task"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupTaskTest(t *testing.T) (*task.Service, *auth.LLMSettingService, *gorm.DB) {
	gormDB, err := db.NewGormDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Manually create tables to avoid ENUM issue in SQLite
	gormDB.Exec(`CREATE TABLE tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        status TEXT NOT NULL DEFAULT 'todo',
        priority TEXT DEFAULT 'medium',
        is_focus_today BOOLEAN DEFAULT FALSE,
        due_at DATETIME,
        created_from TEXT,
        created_at DATETIME,
        updated_at DATETIME
    )`)

	gormDB.Exec(`CREATE TABLE task_steps (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        order_index INTEGER NOT NULL DEFAULT 0,
        title VARCHAR(255) NOT NULL,
        detail TEXT,
        status TEXT NOT NULL DEFAULT 'todo',
        blocking_reason TEXT,
        estimate_minutes INTEGER,
        planned_start DATETIME,
        planned_end DATETIME,
        created_at DATETIME,
        updated_at DATETIME
    )`)

	err = gormDB.AutoMigrate(&auth.UserLLMSetting{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	taskRepo := task.NewRepository(gormDB)
	llmClient := &mockLLMClient{}
	taskSvc := task.NewService(taskRepo, llmClient)

	llmSettingRepo := auth.NewLLMSettingRepository(gormDB)
	llmSettingSvc := auth.NewLLMSettingService(llmSettingRepo, "12345678901234567890123456789012")

	return taskSvc, llmSettingSvc, gormDB
}

func TestTaskHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	taskSvc, llmSettingSvc, _ := setupTaskTest(t)
	handler := internalHTTP.NewTaskHandler(taskSvc, llmSettingSvc)

	router := gin.Default()
	authGroup := router.Group("/api")
	authGroup.Use(func(c *gin.Context) {
		c.Set("userID", uint64(1))
		c.Next()
	})
	handler.RegisterRoutes(authGroup)

	// Setup LLM settings for user 1
	llmSettingSvc.UpdateLLMSetting(nil, 1, auth.LLMSettingRequest{
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "sk-test",
		Model:   "gpt-4",
	})

	t.Run("create from text", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.CreateFromTextReq{
			RawText: "Create a test task with one step",
		})
		req, _ := http.NewRequest("POST", "/api/tasks/from-text", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("list tasks", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/tasks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		tasks := resp["tasks"].([]interface{})
		if len(tasks) == 0 {
			t.Error("expected at least one task")
		}
	})

	t.Run("get task", func(t *testing.T) {
		// First get the task ID from list
		req, _ := http.NewRequest("GET", "/api/tasks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		tasks := resp["tasks"].([]interface{})
		task0 := tasks[0].(map[string]interface{})
		taskID := uint64(task0["id"].(float64))

		req, _ = http.NewRequest("GET", "/api/tasks/"+strconv.FormatUint(taskID, 10), nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("create task", func(t *testing.T) {
		reqBody, _ := json.Marshal(task.Task{
			Title: "Manual Task",
		})
		req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("patch task", func(t *testing.T) {
		title := "Updated Title"
		reqBody, _ := json.Marshal(task.UpdateTaskFields{
			Title: &title,
		})
		req, _ := http.NewRequest("PATCH", "/api/tasks/1", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}
