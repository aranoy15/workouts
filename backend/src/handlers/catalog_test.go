package handlers

import (
	"net/http"
	"strings"
	"testing"
	"workouts-backend/src/models"
	"workouts-backend/src/testhelper"
)

func TestCatalogHandler(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, adminToken, userToken := setupExerciseRouter(t, db, nil)

	t.Run("list muscle groups public", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodGet, "/api/muscle-groups", "", "")
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "legs") {
			t.Errorf("expected legs in body, got %q", w.Body.String())
		}
	})

	t.Run("list levels public", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodGet, "/api/levels", "", "")
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "beginner") {
			t.Errorf("expected beginner in body, got %q", w.Body.String())
		}
	})

	t.Run("create muscle group requires admin", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/muscle-groups", userToken, `{"name":"back"}`)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusForbidden, w.Body.String())
		}
	})

	t.Run("create muscle group", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/muscle-groups", adminToken, `{"name":"back"}`)
		if w.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		}
		var group models.MuscleGroup
		if err := db.Where("name = ?", "back").First(&group).Error; err != nil {
			t.Fatalf("muscle group should exist: %v", err)
		}
	})

	t.Run("create level", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/levels", adminToken, `{"name":"advanced"}`)
		if w.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		}
	})

	t.Run("duplicate muscle group", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/muscle-groups", adminToken, `{"name":"legs"}`)
		if w.Code != http.StatusConflict {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusConflict, w.Body.String())
		}
	})
}
