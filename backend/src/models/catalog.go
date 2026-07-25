package models

import (
	"os"
	"time"
)

type MuscleGroup struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (MuscleGroup) TableName() string {
	if os.Getenv("ENV") == "testsuite" {
		return "muscle_groups"
	}
	return "workouts.muscle_groups"
}

type Level struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Level) TableName() string {
	if os.Getenv("ENV") == "testsuite" {
		return "levels"
	}
	return "workouts.levels"
}
