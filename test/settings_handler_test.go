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

	"github.com/gin-gonic/gin"
)

func setupSettingsTest(t *testing.T) *auth.LLMSettingService {
	gormDB, err := db.NewGormDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = gormDB.AutoMigrate(&auth.UserLLMSetting{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	repo := auth.NewLLMSettingRepository(gormDB)
	// Use a 32-byte key for AES-256
	return auth.NewLLMSettingService(repo, "12345678901234567890123456789012")
}

func TestSettingsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := setupSettingsTest(t)
	handler := internalHTTP.NewSettingsHandler(svc)

	router := gin.Default()
	authGroup := router.Group("/api")
	// Mock userID in context
	authGroup.Use(func(c *gin.Context) {
		c.Set("userID", uint64(1))
		c.Next()
	})
	handler.RegisterRoutes(authGroup)

	t.Run("update settings", func(t *testing.T) {
		reqBody, _ := json.Marshal(auth.LLMSettingRequest{
			BaseURL: "https://api.openai.com/v1",
			APIKey:  "sk-test-key",
			Model:   "gpt-4",
		})
		req, _ := http.NewRequest("POST", "/api/settings/llm", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get settings", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/settings/llm", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["exists"] != true {
			t.Error("expected exists to be true")
		}
		config := resp["config"].(map[string]interface{})
		if config["model"] != "gpt-4" {
			t.Errorf("expected model gpt-4, got %v", config["model"])
		}
	})

	t.Run("delete settings", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/settings/llm", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Verify deletion
		req, _ = http.NewRequest("GET", "/api/settings/llm", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["exists"] != false {
			t.Error("expected exists to be false after deletion")
		}
	})
}
