package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"workouts-backend/src/middleware"
	"workouts-backend/src/models"
	"workouts-backend/src/testhelper"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const testExerciseID = "550e8400-e29b-41d4-a716-446655440010"

type mockS3 struct {
	uploaded map[string][]byte
	deleted  []string
	failUp   bool
	failDel  bool
}

func newMockS3() *mockS3 {
	return &mockS3{uploaded: make(map[string][]byte)}
}

func (m *mockS3) UploadFile(_ context.Context, objectID, key string, body io.Reader, _ string) (string, error) {
	if m.failUp {
		return "", io.ErrUnexpectedEOF
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	objectKey := objectID + "/" + key
	m.uploaded[objectKey] = data
	return "https://storage.yandexcloud.net/workouts-videos/" + objectKey, nil
}

func (m *mockS3) DeleteFile(_ context.Context, key string) error {
	if m.failDel {
		return io.ErrUnexpectedEOF
	}
	m.deleted = append(m.deleted, key)
	delete(m.uploaded, key)
	return nil
}

func (m *mockS3) Bucket() string { return "workouts-videos" }

func setupExerciseRouter(t *testing.T, db *gorm.DB, s3Client S3Service) (*gin.Engine, string, string) {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	testhelper.MustCreateUser(t, db, &models.User{
		ID:       testhelper.TestAdminID,
		Username: "admin_user",
		Email:    "admin@test.local",
		Password: string(hashed),
		Role:     models.UserRoleAdmin,
		IsActive: true,
	})
	testhelper.MustCreateUser(t, db, &models.User{
		ID:       testhelper.TestUserID,
		Username: "regular_user",
		Email:    "user@test.local",
		Password: string(hashed),
		Role:     models.UserRoleUser,
		IsActive: true,
	})

	testhelper.MustCreateMuscleGroup(t, db, &models.MuscleGroup{
		ID:   testhelper.TestMuscleGroupID,
		Name: "legs",
	})
	testhelper.MustCreateLevel(t, db, &models.Level{
		ID:   testhelper.TestLevelID,
		Name: "beginner",
	})

	mgID := testhelper.TestMuscleGroupID
	levelID := testhelper.TestLevelID
	testhelper.MustCreateExercise(t, db, &models.Exercise{
		ID:            testExerciseID,
		Name:          "Squat",
		Description:   "Basic squat",
		MuscleGroupID: &mgID,
		LevelID:       &levelID,
		VideoURLs:     []string{"https://example.com/squat"},
	})

	cfg := testhelper.TestConfig()
	r := gin.New()
	api := r.Group("/api")
	RegisterExerciseHandler(api, db)
	RegisterCatalogHandler(api, db)

	admin := api.Group("")
	middleware.Auth(admin, cfg, db, string(models.UserRoleAdmin))
	RegisterAdminExerciseHandler(admin, db, s3Client)
	RegisterAdminCatalogHandler(admin, db)

	adminToken := testhelper.MustSignJWT(t, testhelper.TestAdminID, cfg.JWTSecret)
	userToken := testhelper.MustSignJWT(t, testhelper.TestUserID, cfg.JWTSecret)
	return r, adminToken, userToken
}

func multipartVideoRequest(t *testing.T, urlPath, token, filename, content string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("video", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, urlPath, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("X-Auth-Token", token)
	}
	return req
}

func TestExerciseHandler_GetExercises(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, _, _ := setupExerciseRouter(t, db, nil)

	w := doJSONRequest(r, http.MethodGet, "/api/exercises", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data []models.Exercise `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 exercise, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "Squat" {
		t.Errorf("name: got %q, want %q", resp.Data[0].Name, "Squat")
	}
}

func TestExerciseHandler_GetExercise(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, _, _ := setupExerciseRouter(t, db, nil)

	t.Run("success", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodGet, "/api/exercises/"+testExerciseID, "", "")
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Squat") {
			t.Errorf("expected Squat in body, got %q", w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodGet, "/api/exercises/00000000-0000-0000-0000-000000000099", "", "")
		if w.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusNotFound, w.Body.String())
		}
	})
}

func TestExerciseHandler_AdminOnly(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	s3 := newMockS3()
	r, _, userToken := setupExerciseRouter(t, db, s3)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		token      string
		wantStatus int
		contains   string
	}{
		{
			name:       "CreateExercise without token",
			method:     http.MethodPost,
			path:       "/api/exercises",
			body:       `{"name":"Deadlift"}`,
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "CreateExercise with user role",
			method:     http.MethodPost,
			path:       "/api/exercises",
			body:       `{"name":"Deadlift"}`,
			token:      userToken,
			wantStatus: http.StatusForbidden,
			contains:   "Forbidden",
		},
		{
			name:       "UpdateExercise without token",
			method:     http.MethodPut,
			path:       "/api/exercises/" + testExerciseID,
			body:       `{"name":"Updated Squat"}`,
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "UpdateExercise with user role",
			method:     http.MethodPut,
			path:       "/api/exercises/" + testExerciseID,
			body:       `{"name":"Updated Squat"}`,
			token:      userToken,
			wantStatus: http.StatusForbidden,
			contains:   "Forbidden",
		},
		{
			name:       "DeleteExercise without token",
			method:     http.MethodDelete,
			path:       "/api/exercises/" + testExerciseID,
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "DeleteExercise with user role",
			method:     http.MethodDelete,
			path:       "/api/exercises/" + testExerciseID,
			token:      userToken,
			wantStatus: http.StatusForbidden,
			contains:   "Forbidden",
		},
		{
			name:       "UploadVideo without token",
			method:     http.MethodPost,
			path:       "/api/videos",
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "DeleteVideo without token",
			method:     http.MethodDelete,
			path:       "/api/videos?key=videos/x/a.mp4",
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "DeleteVideo with user role",
			method:     http.MethodDelete,
			path:       "/api/videos?key=videos/x/a.mp4",
			token:      userToken,
			wantStatus: http.StatusForbidden,
			contains:   "Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "UploadVideo without token" {
				req := multipartVideoRequest(t, tt.path, tt.token, "clip.mp4", "data")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != tt.wantStatus {
					t.Errorf("status: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
				}
				return
			}
			if strings.HasPrefix(tt.name, "UploadVideo with user") {
				req := multipartVideoRequest(t, tt.path, tt.token, "clip.mp4", "data")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != tt.wantStatus {
					t.Errorf("status: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
				}
				return
			}

			w := doJSONRequest(r, tt.method, tt.path, tt.token, tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.contains != "" && !strings.Contains(w.Body.String(), tt.contains) {
				t.Errorf("body should contain %q, got %q", tt.contains, w.Body.String())
			}
		})
	}

	t.Run("UploadVideo with user role", func(t *testing.T) {
		req := multipartVideoRequest(t, "/api/videos", userToken, "clip.mp4", "data")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusForbidden)
		}
	})
}

func TestExerciseHandler_CreateExercise(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, adminToken, _ := setupExerciseRouter(t, db, nil)

	t.Run("success", func(t *testing.T) {
		chest := &models.MuscleGroup{ID: "550e8400-e29b-41d4-a716-446655440011", Name: "chest"}
		intermediate := &models.Level{ID: "550e8400-e29b-41d4-a716-446655440012", Name: "intermediate"}
		testhelper.MustCreateMuscleGroup(t, db, chest)
		testhelper.MustCreateLevel(t, db, intermediate)

		w := doJSONRequest(r, http.MethodPost, "/api/exercises", adminToken,
			`{"name":"Bench Press","description":"Chest press","muscle_group_id":"`+chest.ID+`","level_id":"`+intermediate.ID+`","video_urls":["https://example.com/bench"]}`)
		if w.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var u models.Exercise
		if err := db.Where("name = ?", "Bench Press").First(&u).Error; err != nil {
			t.Fatalf("exercise should exist in DB: %v", err)
		}
		if u.MuscleGroupID == nil || *u.MuscleGroupID != chest.ID {
			t.Errorf("muscle_group_id: got %v, want %q", u.MuscleGroupID, chest.ID)
		}
		if len(u.VideoURLs) != 1 || u.VideoURLs[0] != "https://example.com/bench" {
			t.Errorf("video_urls: got %#v", u.VideoURLs)
		}
	})

	t.Run("invalid muscle_group_id", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/exercises", adminToken,
			`{"name":"Bad","muscle_group_id":"550e8400-e29b-41d4-a716-446655440099"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/exercises", adminToken, `{"description":"no name"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})
}

func TestExerciseHandler_UpdateExercise(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, adminToken, _ := setupExerciseRouter(t, db, nil)

	t.Run("success", func(t *testing.T) {
		quads := &models.MuscleGroup{ID: "550e8400-e29b-41d4-a716-446655440013", Name: "quads"}
		intermediate := &models.Level{ID: "550e8400-e29b-41d4-a716-446655440014", Name: "intermediate"}
		testhelper.MustCreateMuscleGroup(t, db, quads)
		testhelper.MustCreateLevel(t, db, intermediate)

		w := doJSONRequest(r, http.MethodPut, "/api/exercises/"+testExerciseID, adminToken,
			`{"name":"Front Squat","level_id":"`+intermediate.ID+`","muscle_group_id":"`+quads.ID+`","video_urls":["https://example.com/a","https://example.com/b"]}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var exercise models.Exercise
		if err := db.Where("id = ?", testExerciseID).First(&exercise).Error; err != nil {
			t.Fatalf("exercise should exist in DB: %v", err)
		}
		if exercise.Name != "Front Squat" {
			t.Errorf("name: got %q, want %q", exercise.Name, "Front Squat")
		}
		if exercise.LevelID == nil || *exercise.LevelID != intermediate.ID {
			t.Errorf("level_id: got %v, want %q", exercise.LevelID, intermediate.ID)
		}
		if exercise.MuscleGroupID == nil || *exercise.MuscleGroupID != quads.ID {
			t.Errorf("muscle_group_id: got %v, want %q", exercise.MuscleGroupID, quads.ID)
		}
		if exercise.Description != "Basic squat" {
			t.Errorf("description should stay unchanged, got %q", exercise.Description)
		}
		if len(exercise.VideoURLs) != 2 {
			t.Fatalf("video_urls len: got %d, want 2", len(exercise.VideoURLs))
		}
	})

	t.Run("empty name", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/exercises/"+testExerciseID, adminToken, `{"name":""}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/exercises/550e8400-e29b-41d4-a716-446655440099", adminToken,
			`{"name":"Missing"}`)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusNotFound, w.Body.String())
		}
	})
}

func TestExerciseHandler_DeleteExercise(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, adminToken, _ := setupExerciseRouter(t, db, nil)

	t.Run("success", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodDelete, "/api/exercises/"+testExerciseID, adminToken, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var e models.Exercise
		if err := db.Where("id = ?", testExerciseID).First(&e).Error; err == nil {
			t.Fatal("exercise should be soft-deleted")
		}

		w = doJSONRequest(r, http.MethodGet, "/api/exercises/"+testExerciseID, "", "")
		if w.Code != http.StatusNotFound {
			t.Fatalf("deleted exercise should return 404, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodDelete, "/api/exercises/00000000-0000-0000-0000-000000000099", adminToken, "")
		if w.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusNotFound, w.Body.String())
		}
	})
}

func TestExerciseHandler_UploadVideo(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	s3 := newMockS3()
	r, adminToken, _ := setupExerciseRouter(t, db, s3)

	t.Run("success", func(t *testing.T) {
		req := multipartVideoRequest(t, "/api/videos", adminToken, "squat.mp4", "video-bytes")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var resp struct {
			Data struct {
				Key      string `json:"key"`
				VideoURL string `json:"video_url"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !strings.HasPrefix(resp.Data.Key, "videos/") || !strings.HasSuffix(resp.Data.Key, "/squat.mp4") {
			t.Errorf("unexpected key: %q", resp.Data.Key)
		}
		wantURL := "https://storage.yandexcloud.net/workouts-videos/" + resp.Data.Key
		if resp.Data.VideoURL != wantURL {
			t.Errorf("video_url: got %q, want %q", resp.Data.VideoURL, wantURL)
		}
		if _, ok := s3.uploaded[resp.Data.Key]; !ok {
			t.Errorf("expected upload in mock for key %q", resp.Data.Key)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		_ = mw.WriteField("other", "x")
		_ = mw.Close()
		req := httptest.NewRequest(http.MethodPost, "/api/videos", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("X-Auth-Token", adminToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})
}

func TestExerciseHandler_DeleteVideo(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	s3 := newMockS3()
	r, adminToken, _ := setupExerciseRouter(t, db, s3)
	key := "videos/550e8400-e29b-41d4-a716-446655440099/squat.mp4"
	s3.uploaded[key] = []byte("x")

	t.Run("by key query", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodDelete, "/api/videos?key="+key, adminToken, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if len(s3.deleted) == 0 || s3.deleted[len(s3.deleted)-1] != key {
			t.Errorf("expected deleted key %q, got %v", key, s3.deleted)
		}
	})

	t.Run("by video_url json", func(t *testing.T) {
		s3.uploaded[key] = []byte("x")
		body := `{"video_url":"https://storage.yandexcloud.net/workouts-videos/` + key + `"}`
		req := httptest.NewRequest(http.MethodDelete, "/api/videos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Auth-Token", adminToken)
		req.ContentLength = int64(len(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("missing key and url", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodDelete, "/api/videos", adminToken, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})

	t.Run("invalid key prefix", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodDelete, "/api/videos?key=other/file.mp4", adminToken, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})
}
