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

func setupTestDB(t *testing.T) *auth.Service {
	gormDB, err := db.NewGormDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = gormDB.AutoMigrate(&auth.User{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	jwtMgr := auth.NewJWTManager("test-secret")
	return auth.NewService(gormDB, jwtMgr)
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := setupTestDB(t)
	handler := internalHTTP.NewAuthHandler(svc)

	router := gin.Default()
	api := router.Group("/api")
	handler.RegisterRoutes(api.Group("/auth"))

	t.Run("successful registration", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.RegisterReq{
			Email:    "test@example.com",
			Password: "password123",
		})
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.RegisterReq{
			Email:    "invalid-email",
			Password: "password123",
		})
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("password too short", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.RegisterReq{
			Email:    "test2@example.com",
			Password: "123",
		})
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := setupTestDB(t)
	handler := internalHTTP.NewAuthHandler(svc)

	router := gin.Default()
	api := router.Group("/api")
	handler.RegisterRoutes(api.Group("/auth"))

	// Pre-register a user
	svc.Register(nil, "test@example.com", "password123")

	t.Run("successful login", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.LoginReq{
			Email:    "test@example.com",
			Password: "password123",
		})
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["token"] == "" {
			t.Error("expected token in response")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		reqBody, _ := json.Marshal(internalHTTP.LoginReq{
			Email:    "test@example.com",
			Password: "wrongpassword",
		})
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})
}
