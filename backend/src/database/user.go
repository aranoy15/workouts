package database

import (
	"errors"
	"log"
	"strings"
	"workouts-backend/src/models"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func CreateUser(db *gorm.DB, user *models.User) error {
	q := db
	if user.Email == "" {
		q = db.Omit("email")
	}
	if err := q.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
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

func GetUserByUsername(db *gorm.DB, username string) (*models.User, error) {
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("User %s not found", username)
			return nil, ErrUserNotFound
		}
		log.Printf("Error getting user %s: %v", username, err)
		return nil, err
	}
	return &user, nil
}

func GetUsers(db *gorm.DB) ([]models.User, error) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Printf("Error getting users: %v", err)
		return nil, err
	}
	return users, nil
}

func CountUsers(db *gorm.DB) (int64, error) {
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		log.Printf("Error counting users: %v", err)
		return 0, err
	}
	return count, nil
}

func UpdateUser(db *gorm.DB, user *models.User) error {
	updates := map[string]interface{}{
		"username":  user.Username,
		"password":  user.Password,
		"role":      user.Role,
		"is_active": user.IsActive,
	}
	if user.Email == "" {
		updates["email"] = nil
	} else {
		updates["email"] = user.Email
	}

	if err := db.Model(user).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Printf("User %s already exists", user.ID)
			return ErrUserAlreadyExists
		}
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
