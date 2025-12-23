package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"assistant-qisumi/internal/agent"
	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/db"
	"assistant-qisumi/internal/dependency"
	internalHTTP "assistant-qisumi/internal/http"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupSessionTest(t *testing.T) (*agent.Service, *session.Repository, *auth.LLMSettingService, *gorm.DB) {
	gormDB, err := db.NewGormDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	gormDB.Exec(`CREATE TABLE sessions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        task_id INTEGER,
        type TEXT NOT NULL DEFAULT "task",
        created_at DATETIME
    )`)

	gormDB.Exec(`CREATE TABLE messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id INTEGER NOT NULL,
        role TEXT NOT NULL,
        agent_name VARCHAR(64),
        content TEXT NOT NULL,
        created_at DATETIME
    )`)

	gormDB.Exec(`CREATE TABLE tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        status TEXT NOT NULL DEFAULT "todo",
        priority TEXT DEFAULT "medium",
        is_focus_today BOOLEAN DEFAULT FALSE,
        due_at DATETIME,
        created_from TEXT,
        created_at DATETIME,
        updated_at DATETIME
    )`)

	err = gormDB.AutoMigrate(&auth.UserLLMSetting{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	taskRepo := task.NewRepository(gormDB)
	sessionRepo := session.NewRepository(gormDB)
	dependencySvc := dependency.NewService(gormDB, taskRepo, sessionRepo)

	router := &mockRouter{}
	agents := []agent.Agent{&mockAgent{}}
	llmClient := &mockLLMClient{}

	agentSvc := agent.NewService(router, agents, taskRepo, sessionRepo, dependencySvc, gormDB, llmClient)

	llmSettingRepo := auth.NewLLMSettingRepository(gormDB)
	llmSettingSvc := auth.NewLLMSettingService(llmSettingRepo, "12345678901234567890123456789012", nil)

	return agentSvc, sessionRepo, llmSettingSvc, gormDB
}

func TestSessionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	agentSvc, sessionRepo, llmSettingSvc, _ := setupSessionTest(t)
	handler := internalHTTP.NewSessionHandler(agentSvc, sessionRepo, llmSettingSvc)

	router := gin.Default()
	authGroup := router.Group("/api")
	authGroup.Use(func(c *gin.Context) {
		c.Set("userID", uint64(1))
		c.Next()
	})
	handler.RegisterRoutes(authGroup)

	t.Run("invalid session id", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.PostMessageReq{
			Content: "Hello",
		})
		req, _ := http.NewRequest("POST", "/api/sessions/abc/messages", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing llm config", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.PostMessageReq{
			Content: "Hello",
		})
		req, _ := http.NewRequest("POST", "/api/sessions/1/messages", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d, body: %s", w.Code, w.Body.String())
		}
	})
}
