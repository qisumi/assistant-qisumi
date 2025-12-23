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
	return auth.NewLLMSettingService(repo, "12345678901234567890123456789012", nil)
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

	t.Run("get settings empty", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/settings/llm", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		var resp auth.LLMConfig
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.BaseURL != "" {
			t.Error("expected empty config when default is nil")
		}
	})

	t.Run("update settings invalid body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/settings/llm", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
