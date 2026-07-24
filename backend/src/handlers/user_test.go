package handlers

import (
	"bytes"
	"encoding/json"
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

func setupUserRouter(t *testing.T, db *gorm.DB) (*gin.Engine, string, string) {
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

	cfg := testhelper.TestConfig()
	r := gin.New()
	api := r.Group("/api")
	RegisterUserHandler(api, cfg, db)

	admin := api.Group("")
	middleware.Auth(admin, cfg, db, string(models.UserRoleAdmin))
	RegisterAdminUserHandler(admin, cfg, db)

	adminToken := testhelper.MustSignJWT(t, testhelper.TestAdminID, cfg.JWTSecret)
	userToken := testhelper.MustSignJWT(t, testhelper.TestUserID, cfg.JWTSecret)
	return r, adminToken, userToken
}

func doJSONRequest(r http.Handler, method, path, token string, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUserHandler_Login(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, _, _ := setupUserRouter(t, db)

	t.Run("success by username", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", `{"username":"regular_user","password":"secret"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"token"`) {
			t.Errorf("expected token in body, got %q", w.Body.String())
		}
	})

	t.Run("success by email", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", `{"email":"admin@test.local","password":"secret"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"token"`) {
			t.Errorf("expected token in body, got %q", w.Body.String())
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", `{"username":"regular_user","password":"wrong"}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusUnauthorized, w.Body.String())
		}
	})

	t.Run("user not found", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", `{"username":"missing","password":"secret"}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusUnauthorized, w.Body.String())
		}
	})

	t.Run("missing password", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", `{"username":"regular_user"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})

	t.Run("missing email and username", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", `{"password":"secret"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})

	t.Run("inactive user", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		inactiveID := "550e8400-e29b-41d4-a716-446655440099"
		testhelper.MustCreateUser(t, db, &models.User{
			ID:       inactiveID,
			Username: "inactive_user",
			Email:    "inactive@test.local",
			Password: string(hashed),
			Role:     models.UserRoleUser,
			IsActive: true,
		})
		if err := db.Model(&models.User{}).Where("id = ?", inactiveID).Update("is_active", false).Error; err != nil {
			t.Fatalf("deactivate user: %v", err)
		}
		w := doJSONRequest(r, http.MethodPost, "/api/auth/login", "", `{"username":"inactive_user","password":"secret"}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusUnauthorized, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Invalid username or password") {
			t.Errorf("expected generic auth error, got %q", w.Body.String())
		}
	})
}

func TestUserHandler_AdminOnly(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, _, userToken := setupUserRouter(t, db)

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
			name:       "CreateUser without token",
			method:     http.MethodPost,
			path:       "/api/users",
			body:       `{"username":"x","password":"secret"}`,
			token:      "",
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "CreateUser with user role",
			method:     http.MethodPost,
			path:       "/api/users",
			body:       `{"username":"x","password":"secret"}`,
			token:      userToken,
			wantStatus: http.StatusForbidden,
			contains:   "Forbidden",
		},
		{
			name:       "GetUsers without token",
			method:     http.MethodGet,
			path:       "/api/users",
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "GetUsers with user role",
			method:     http.MethodGet,
			path:       "/api/users",
			token:      userToken,
			wantStatus: http.StatusForbidden,
			contains:   "Forbidden",
		},
		{
			name:       "UpdateUser without token",
			method:     http.MethodPut,
			path:       "/api/users/" + testhelper.TestUserID,
			body:       `{"username":"new_name"}`,
			wantStatus: http.StatusUnauthorized,
			contains:   "Unauthorized",
		},
		{
			name:       "UpdateUser with user role",
			method:     http.MethodPut,
			path:       "/api/users/" + testhelper.TestUserID,
			body:       `{"username":"new_name"}`,
			token:      userToken,
			wantStatus: http.StatusForbidden,
			contains:   "Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doJSONRequest(r, tt.method, tt.path, tt.token, tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.contains != "" && !strings.Contains(w.Body.String(), tt.contains) {
				t.Errorf("body should contain %q, got %q", tt.contains, w.Body.String())
			}
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, adminToken, _ := setupUserRouter(t, db)

	t.Run("success", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/users", adminToken,
			`{"username":"new_user","email":"new@test.local","password":"secret123","role":"user"}`)
		if w.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var resp models.Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		var u models.User
		if err := db.Where("username = ?", "new_user").First(&u).Error; err != nil {
			t.Fatalf("user should exist in DB: %v", err)
		}
		if u.Password == "secret123" {
			t.Error("password must be hashed")
		}
		if u.Role != models.UserRoleUser {
			t.Errorf("role: got %q, want %q", u.Role, models.UserRoleUser)
		}
		if !u.IsActive {
			t.Error("expected active user")
		}
	})

	t.Run("default role user", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/users", adminToken,
			`{"username":"default_role","password":"secret123"}`)
		if w.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		}
		var u models.User
		if err := db.Where("username = ?", "default_role").First(&u).Error; err != nil {
			t.Fatalf("user should exist: %v", err)
		}
		if u.Role != models.UserRoleUser {
			t.Errorf("role: got %q, want %q", u.Role, models.UserRoleUser)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/users", adminToken,
			`{"username":"bad_role","password":"secret123","role":"superadmin"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/users", adminToken,
			`{"username":"regular_user","password":"secret123"}`)
		if w.Code != http.StatusConflict {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusConflict, w.Body.String())
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPost, "/api/users", adminToken, `{"username":"no_pass"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, adminToken, _ := setupUserRouter(t, db)

	t.Run("success", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/users/"+testhelper.TestUserID, adminToken,
			`{"username":"updated_user","email":"updated@test.local"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var u models.User
		if err := db.Where("id = ?", testhelper.TestUserID).First(&u).Error; err != nil {
			t.Fatalf("get user: %v", err)
		}
		if u.Username != "updated_user" {
			t.Errorf("username: got %q, want %q", u.Username, "updated_user")
		}
		if u.Email != "updated@test.local" {
			t.Errorf("email: got %q, want %q", u.Email, "updated@test.local")
		}
	})

	t.Run("update role and is_active", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/users/"+testhelper.TestUserID, adminToken,
			`{"role":"admin","is_active":false}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var u models.User
		if err := db.Where("id = ?", testhelper.TestUserID).First(&u).Error; err != nil {
			t.Fatalf("get user: %v", err)
		}
		if u.Role != models.UserRoleAdmin {
			t.Errorf("role: got %q, want %q", u.Role, models.UserRoleAdmin)
		}
		if u.IsActive {
			t.Error("expected inactive user")
		}
	})

	t.Run("update password", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/users/"+testhelper.TestAdminID, adminToken,
			`{"password":"newsecret"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var u models.User
		if err := db.Where("id = ?", testhelper.TestAdminID).First(&u).Error; err != nil {
			t.Fatalf("get user: %v", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("newsecret")); err != nil {
			t.Error("password was not updated correctly")
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/users/00000000-0000-0000-0000-000000000099", adminToken,
			`{"username":"nope"}`)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusNotFound, w.Body.String())
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/users/"+testhelper.TestAdminID, adminToken,
			`{"role":"superadmin"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})

	t.Run("empty username", func(t *testing.T) {
		w := doJSONRequest(r, http.MethodPut, "/api/users/"+testhelper.TestAdminID, adminToken,
			`{"username":""}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})
}

func TestUserHandler_GetUsers(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r, adminToken, _ := setupUserRouter(t, db)

	w := doJSONRequest(r, http.MethodGet, "/api/users", adminToken, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data []models.User `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data) < 2 {
		t.Fatalf("expected at least 2 users, got %d", len(resp.Data))
	}
	for _, u := range resp.Data {
		if u.Password != "" {
			t.Errorf("password must not be exposed for user %s", u.ID)
		}
	}
}
