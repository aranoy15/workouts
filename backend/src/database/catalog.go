package database

import (
	"errors"
	"log"
	"strings"
	"workouts-backend/src/models"

	"gorm.io/gorm"
)

var (
	ErrMuscleGroupNotFound      = errors.New("muscle group not found")
	ErrMuscleGroupAlreadyExists = errors.New("muscle group already exists")
	ErrLevelNotFound            = errors.New("level not found")
	ErrLevelAlreadyExists       = errors.New("level already exists")
)

func CreateMuscleGroup(db *gorm.DB, group *models.MuscleGroup) error {
	if err := db.Create(group).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrMuscleGroupAlreadyExists
		}
		log.Printf("Error creating muscle group %s: %v", group.ID, err)
		return err
	}
	return nil
}

func GetMuscleGroupByID(db *gorm.DB, id string) (*models.MuscleGroup, error) {
	var group models.MuscleGroup
	if err := db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMuscleGroupNotFound
		}
		log.Printf("Error getting muscle group %s: %v", id, err)
		return nil, err
	}
	return &group, nil
}

func GetMuscleGroups(db *gorm.DB) ([]models.MuscleGroup, error) {
	var groups []models.MuscleGroup
	if err := db.Order("name asc").Find(&groups).Error; err != nil {
		log.Printf("Error getting muscle groups: %v", err)
		return nil, err
	}
	return groups, nil
}

func CreateLevel(db *gorm.DB, level *models.Level) error {
	if err := db.Create(level).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrLevelAlreadyExists
		}
		log.Printf("Error creating level %s: %v", level.ID, err)
		return err
	}
	return nil
}

func GetLevelByID(db *gorm.DB, id string) (*models.Level, error) {
	var level models.Level
	if err := db.Where("id = ?", id).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLevelNotFound
		}
		log.Printf("Error getting level %s: %v", id, err)
		return nil, err
	}
	return &level, nil
}

func GetLevels(db *gorm.DB) ([]models.Level, error) {
	var levels []models.Level
	if err := db.Order("name asc").Find(&levels).Error; err != nil {
		log.Printf("Error getting levels: %v", err)
		return nil, err
	}
	return levels, nil
}
