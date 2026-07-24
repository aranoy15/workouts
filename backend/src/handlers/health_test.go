package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"workouts-backend/src/testhelper"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestHealthHandler_CheckHealth(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	handler := NewHealthHandler(db)
	router := gin.New()
	router.GET("/health", handler.CheckHealth)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["message"] != "OK" {
		t.Errorf("message: got %q, want %q", body["message"], "OK")
	}
}

func TestRegisterHealthHandler(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	router := gin.New()
	api := router.Group("/api")
	RegisterHealthHandler(api, db)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["message"] != "OK" {
		t.Errorf("message: got %q, want %q", body["message"], "OK")
	}
}
