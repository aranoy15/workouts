package models

import (
	"os"
	"time"

	"gorm.io/gorm"
)

type Exercise struct {
	ID            string         `json:"id" gorm:"primaryKey;type:uuid"`
	Name          string         `json:"name" gorm:"not null"`
	Description   string         `json:"description"`
	MuscleGroupID *string        `json:"muscle_group_id" gorm:"type:uuid"`
	LevelID       *string        `json:"level_id" gorm:"type:uuid"`
	MuscleGroup   *MuscleGroup   `json:"muscle_group,omitempty" gorm:"foreignKey:MuscleGroupID"`
	Level         *Level         `json:"level,omitempty" gorm:"foreignKey:LevelID"`
	VideoURLs     []string       `json:"video_urls" gorm:"serializer:json"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Exercise) TableName() string {
	if os.Getenv("ENV") == "testsuite" {
		return "exercises"
	}
	return "workouts.exercises"
}
