package database

import (
	"errors"
	"log"
	"workouts-backend/src/models"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func CreateUser(db *gorm.DB, user *models.User) error {
	if err := db.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.Printf("User %s already exists", user.ID)
			return ErrUserAlreadyExists
		}
		log.Printf("Error creating user %s: %v", user.ID, err)
		return err
	}
	return nil
}

func GetUserByID(db *gorm.DB, id string) (*models.User, error) {
	var user models.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("User %s not found", id)
			return nil, ErrUserNotFound
		}
		log.Printf("Error getting user %s: %v", id, err)
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("User %s not found", email)
			return nil, ErrUserNotFound
		}
		log.Printf("Error getting user %s: %v", email, err)
		return nil, err
	}
	return &user, nil
}

func UpdateUser(db *gorm.DB, user *models.User) error {
	if err := db.Save(user).Error; err != nil {
		log.Printf("Error updating user %s: %v", user.ID, err)
		return err
	}
	return nil
}

func DeleteUser(db *gorm.DB, id string) error {
	if err := db.Delete(&models.User{}, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("User %s not found", id)
			return ErrUserNotFound
		}
		log.Printf("Error deleting user %s: %v", id, err)
		return err
	}
	return nil
}
