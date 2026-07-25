package testhelper

import (
	"os"
	"testing"
	"time"
	"workouts-backend/src/config"
	"workouts-backend/src/models"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	TestAdminID = "550e8400-e29b-41d4-a716-446655440001"
	TestUserID  = "550e8400-e29b-41d4-a716-446655440002"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("ENV", "testsuite")
	_ = os.Setenv("ENV", "testsuite")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME,
			updated_at DATETIME,
			is_active INTEGER NOT NULL DEFAULT 1
		)
	`).Error; err != nil {
		t.Fatalf("create users table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE exercises (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			muscle_group TEXT,
			level TEXT,
			video_url TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create exercises table: %v", err)
	}

	return db
}

func MustCreateUser(t *testing.T, db *gorm.DB, u *models.User) {
	t.Helper()
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("create user in test: %v", err)
	}
}

func MustCreateExercise(t *testing.T, db *gorm.DB, e *models.Exercise) {
	t.Helper()
	if err := db.Create(e).Error; err != nil {
		t.Fatalf("create exercise in test: %v", err)
	}
}

func TestConfig() *config.Config {
	return &config.Config{
		Port:      "8080",
		JWTSecret: "test-jwt-secret",
	}
}

func MustSignJWT(t *testing.T, userID, secret string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return token
}
