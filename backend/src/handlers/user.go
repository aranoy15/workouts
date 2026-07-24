package handlers

import (
	"errors"
	"net/http"
	"time"
	"workouts-backend/src/config"
	"workouts-backend/src/database"
	"workouts-backend/src/middleware"
	"workouts-backend/src/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const loginRateLimit = 20

var (
	loginRateWindow    = time.Minute
	dummyPasswordHash  []byte
	invalidCredentials = "Invalid username or password"
)

func init() {
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
	if err != nil {
		panic("failed to generate dummy password hash: " + err.Error())
	}
	dummyPasswordHash = hash
}

type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

type CreateUserRequest struct {
	Username string          `json:"username" binding:"required"`
	Email    string          `json:"email"`
	Password string          `json:"password" binding:"required"`
	Role     models.UserRole `json:"role"`
}

type UpdateUserRequest struct {
	Username *string          `json:"username"`
	Email    *string          `json:"email"`
	Password *string          `json:"password"`
	Role     *models.UserRole `json:"role"`
	IsActive *bool            `json:"is_active"`
}

type UserHandler struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewUserHandler(cfg *config.Config, db *gorm.DB) *UserHandler {
	return &UserHandler{cfg: cfg, db: db}
}

func RegisterUserHandler(router *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {
	h := NewUserHandler(cfg, db)
	router.POST("/auth/login", middleware.RateLimit(loginRateLimit, loginRateWindow), h.Login)
}

func RegisterAdminUserHandler(router *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {
	h := NewUserHandler(cfg, db)
	router.POST("/users", h.CreateUser)
	router.PUT("/users/:id", h.UpdateUser)
	router.GET("/users", h.GetUsers)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
		return
	}
	if req.Email == "" && req.Username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Email or username is required"))
		return
	}

	var user *models.User
	var err error
	if req.Username != "" {
		user, err = database.GetUserByUsername(h.db, req.Username)
	} else {
		user, err = database.GetUserByEmail(h.db, req.Email)
	}
	if err != nil && !errors.Is(err, database.ErrUserNotFound) {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get user"))
		return
	}

	passwordHash := dummyPasswordHash
	active := false
	if err == nil && user != nil && user.IsActive {
		passwordHash = []byte(user.Password)
		active = true
	}

	passwordOK := bcrypt.CompareHashAndPassword(passwordHash, []byte(req.Password)) == nil
	if !active || !passwordOK {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(invalidCredentials))
		return
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    string(user.Role),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}).SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create token"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"token": token,
		"user":  user,
	}))
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
		return
	}

	role := req.Role
	if role == "" {
		role = models.UserRoleUser
	}
	if role != models.UserRoleAdmin && role != models.UserRoleUser {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid role"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to hash password"))
		return
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
		IsActive: true,
	}

	if err := database.CreateUser(h.db, user); err != nil {
		if errors.Is(err, database.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create user"))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(user))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	user, err := database.GetUserByID(h.db, id)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get user"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
		return
	}

	if req.Username != nil {
		if *req.Username == "" {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("Username is required"))
			return
		}
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		if *req.Role != models.UserRoleAdmin && *req.Role != models.UserRoleUser {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid role"))
			return
		}
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to hash password"))
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := database.UpdateUser(h.db, user); err != nil {
		if errors.Is(err, database.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to update user"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(user))
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := database.GetUsers(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get users"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(users))
}
