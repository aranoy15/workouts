package models

import (
	"os"
	"time"

	"gorm.io/gorm"
)

type Exercise struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	MuscleGroup string         `json:"muscle_group"`
	Level       string         `json:"level"`
	VideoURL    string         `json:"video_url"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Exercise) TableName() string {
	if os.Getenv("ENV") == "testsuite" {
		return "exercises"
	}
	return "workouts.exercises"
}
