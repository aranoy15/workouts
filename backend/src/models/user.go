package models

import (
	"os"
	"time"
)

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`
	Username  string    `gorm:"not null" json:"username"`
	Email     string    `gorm:"default:null" json:"email"`
	Password  string    `gorm:"default:null" json:"password"`
	Role      UserRole  `gorm:"default:user" json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

func (User) TableName() string {
	if os.Getenv("ENV") == "testsuite" {
		return "users"
	}
	return "workouts.users"
}
