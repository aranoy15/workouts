package database

import (
	"errors"
	"log"
	"workouts-backend/src/models"

	"gorm.io/gorm"
)

var (
	ErrExerciseNotFound = errors.New("exercise not found")
)

func CreateExercise(db *gorm.DB, exercise *models.Exercise) error {
	if err := db.Create(exercise).Error; err != nil {
		log.Printf("Error creating exercise %s: %v", exercise.ID, err)
		return err
	}
	return nil
}

func GetExerciseByID(db *gorm.DB, id string) (*models.Exercise, error) {
	var exercise models.Exercise
	if err := db.Where("id = ?", id).First(&exercise).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Exercise %s not found", id)
			return nil, ErrExerciseNotFound
		}
		log.Printf("Error getting exercise %s: %v", id, err)
		return nil, err
	}
	return &exercise, nil
}

func GetExercises(db *gorm.DB) ([]models.Exercise, error) {
	var exercises []models.Exercise
	if err := db.Find(&exercises).Error; err != nil {
		log.Printf("Error getting exercises: %v", err)
		return nil, err
	}
	return exercises, nil
}

func DeleteExercise(db *gorm.DB, id string) error {
	result := db.Where("id = ?", id).Delete(&models.Exercise{})
	if result.Error != nil {
		log.Printf("Error deleting exercise %s: %v", id, result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Printf("Exercise %s not found", id)
		return ErrExerciseNotFound
	}
	return nil
}
