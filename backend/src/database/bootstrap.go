package database

import (
	"fmt"
	"log"
	"workouts-backend/src/config"
	"workouts-backend/src/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// EnsureBootstrapAdmin creates the initial admin when the users table is empty.
// Requires ADMIN_PASSWORD (and optionally ADMIN_USERNAME / ADMIN_EMAIL) in config.
func EnsureBootstrapAdmin(db *gorm.DB, cfg *config.Config) error {
	count, err := CountUsers(db)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	if count > 0 {
		return nil
	}

	if cfg.AdminPassword == "" {
		return fmt.Errorf("database has no users; set ADMIN_PASSWORD to create the initial admin")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	admin := &models.User{
		ID:       uuid.New().String(),
		Username: cfg.AdminUsername,
		Email:    cfg.AdminEmail,
		Password: string(hashedPassword),
		Role:     models.UserRoleAdmin,
		IsActive: true,
	}

	if err := CreateUser(db, admin); err != nil {
		return fmt.Errorf("failed to create bootstrap admin: %w", err)
	}

	log.Printf("Bootstrap admin created: username=%s", admin.Username)
	return nil
}
