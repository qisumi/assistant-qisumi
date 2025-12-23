package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/db"
	internalHTTP "assistant-qisumi/internal/http"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupTaskTest(t *testing.T) (*task.Service, *session.Repository, *auth.LLMSettingService, *gorm.DB) {
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

	// 迁移 Session 相关表
	err = gormDB.AutoMigrate(&session.Session{}, &session.Message{})
	if err != nil {
		t.Fatalf("failed to migrate session tables: %v", err)
	}

	err = gormDB.AutoMigrate(&auth.UserLLMSetting{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	taskRepo := task.NewRepository(gormDB)
	sessionRepo := session.NewRepository(gormDB)
	llmClient := &mockLLMClient{}
	taskSvc := task.NewService(taskRepo, llmClient)

	llmSettingRepo := auth.NewLLMSettingRepository(gormDB)
	llmSettingSvc := auth.NewLLMSettingService(llmSettingRepo, "12345678901234567890123456789012", nil)

	return taskSvc, sessionRepo, llmSettingSvc, gormDB
}

func TestTaskHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	taskSvc, sessionRepo, llmSettingSvc, _ := setupTaskTest(t)
	handler := internalHTTP.NewTaskHandler(taskSvc, sessionRepo, llmSettingSvc)

	router := gin.Default()
	authGroup := router.Group("/api")
	authGroup.Use(func(c *gin.Context) {
		c.Set("userID", uint64(1))
		c.Next()
	})
	handler.RegisterRoutes(authGroup)

	t.Run("create from text without llm config", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.CreateFromTextReq{
			RawText: "Create a test task with one step",
		})
		req, _ := http.NewRequest("POST", "/api/tasks/from-text", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("create task invalid json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("get task invalid id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/tasks/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("patch task invalid id", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/api/tasks/abc", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("patch task invalid json", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/api/tasks/1", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
