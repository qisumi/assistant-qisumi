package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/config"
	"assistant-qisumi/internal/db"
	internalHTTP "assistant-qisumi/internal/http"
	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

	"github.com/gin-gonic/gin"
)

func TestAPIIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Setup Test Server with In-Memory SQLite
	gormDB, err := db.NewGormDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(gormDB)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	cfg := config.HTTPConfig{Port: "8080"}
	jwtCfg := config.JWTConfig{Secret: "test-secret", ExpireHour: 24}
	cryptoCfg := config.CryptoConfig{APIKeyEncryptionKey: "12345678901234567890123456789012"}
	llmCfg := config.LLMConfig{
		APIKey:     "test-api-key",
		ModelName:  "test-model",
		APIBaseURL: "https://api.test.com",
	}

	mockLLM := &mockLLMClient{}
	server := internalHTTP.NewServer(cfg, jwtCfg, cryptoCfg, llmCfg, gormDB, mockLLM)

	var token string

	// 2. Health Check
	t.Run("Health Check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 3. Auth Required
	t.Run("Auth Required", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/tasks", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 4. Test Registration
	t.Run("Register User", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.RegisterReq{
			Email:    "test@example.com",
			Password: "password123",
		})
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 5. Test Login
	t.Run("Login User", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.LoginReq{
			Email:    "test@example.com",
			Password: "password123",
		})
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		token = resp["token"]
		if token == "" {
			t.Fatal("expected token in response")
		}
	})

	// 6. Test LLM Settings (Empty)
	t.Run("Get LLM Settings Empty", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/settings/llm", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp auth.LLMConfig
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.BaseURL != "https://api.test.com" {
			t.Errorf("expected default BaseURL, got %s", resp.BaseURL)
		}
		if !resp.IsDefault {
			t.Error("expected IsDefault to be true")
		}
	})

	// 7. Test LLM Settings Update
	t.Run("Update LLM Settings", func(t *testing.T) {
		reqBody, _ := json.Marshal(auth.LLMSettingRequest{
			BaseURL: "https://api.openai.com/v1",
			APIKey:  "sk-test-key",
			Model:   "gpt-3.5-turbo",
		})
		req, _ := http.NewRequest("POST", "/api/settings/llm", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 8. Test LLM Settings Get
	t.Run("Get LLM Settings", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/settings/llm", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp auth.LLMConfig
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.BaseURL == "" {
			t.Fatal("expected non-empty config")
		}
	})

	// 9. Test Task Creation
	var taskID uint64
	t.Run("Create Task", func(t *testing.T) {
		reqBody, _ := json.Marshal(task.Task{
			Title:       "Test Task",
			Description: "Test Description",
			Priority:    "high",
		})
		req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Task task.Task `json:"task"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		taskID = resp.Task.ID
		if taskID == 0 {
			t.Fatal("expected task ID to be non-zero")
		}
	})

	// 10. Test List Tasks
	t.Run("List Tasks", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Tasks []task.Task `json:"tasks"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Tasks) == 0 {
			t.Error("expected at least one task")
		}
	})

	// 11. Test Get Task
	t.Run("Get Task", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/tasks/"+strconv.FormatUint(taskID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Task task.Task `json:"task"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Task.ID != taskID {
			t.Errorf("expected task ID %d, got %d", taskID, resp.Task.ID)
		}
	})

	// 12. Test Patch Task
	t.Run("Patch Task", func(t *testing.T) {
		newTitle := "Updated Task Title"
		reqBody, _ := json.Marshal(task.UpdateTaskFields{
			Title: &newTitle,
		})
		req, _ := http.NewRequest("PATCH", "/api/tasks/"+strconv.FormatUint(taskID, 10), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify update
		req, _ = http.NewRequest("GET", "/api/tasks/"+strconv.FormatUint(taskID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)

		var resp struct {
			Task task.Task `json:"task"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Task.Title != newTitle {
			t.Errorf("expected title %s, got %s", newTitle, resp.Task.Title)
		}
	})

	// 13. Test Create Task From Text
	t.Run("Create Task From Text", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.CreateFromTextReq{
			RawText: "I need to buy milk and eggs",
		})
		req, _ := http.NewRequest("POST", "/api/tasks/from-text", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Task task.Task `json:"task"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Task.Title == "" {
			t.Error("expected task title to be non-empty")
		}
	})

	// 14. Test Session Message
	t.Run("Post Session Message", func(t *testing.T) {
		// Create a session first
		sess := session.Session{
			UserID: 1, // The user we registered
			Type:   "global",
		}
		gormDB.Create(&sess)

		reqBody, _ := json.Marshal(internalHTTP.PostMessageReq{
			Content: "Hello agent",
		})
		req, _ := http.NewRequest("POST", "/api/sessions/"+strconv.FormatUint(sess.ID, 10)+"/messages", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			AssistantMessage string `json:"assistantMessage"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.AssistantMessage == "" {
			t.Error("expected assistant message to be non-empty")
		}
	})

	// 15. Test Delete LLM Settings
	t.Run("Delete LLM Settings", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/settings/llm", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 16. Test LLM Settings After Delete
	t.Run("Get LLM Settings After Delete", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/settings/llm", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp auth.LLMConfig
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.BaseURL != "https://api.test.com" {
			t.Error("expected default config after deletion")
		}
		if !resp.IsDefault {
			t.Error("expected IsDefault to be true")
		}
	})
}
