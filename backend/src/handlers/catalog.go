package handlers

import (
	"errors"
	"net/http"
	"strings"
	"workouts-backend/src/database"
	"workouts-backend/src/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateNamedRequest struct {
	Name string `json:"name" binding:"required"`
}

type CatalogHandler struct {
	db *gorm.DB
}

func NewCatalogHandler(db *gorm.DB) *CatalogHandler {
	return &CatalogHandler{db: db}
}

func RegisterCatalogHandler(router *gin.RouterGroup, db *gorm.DB) {
	h := NewCatalogHandler(db)
	router.GET("/muscle-groups", h.GetMuscleGroups)
	router.GET("/levels", h.GetLevels)
}

func RegisterAdminCatalogHandler(router *gin.RouterGroup, db *gorm.DB) {
	h := NewCatalogHandler(db)
	router.POST("/muscle-groups", h.CreateMuscleGroup)
	router.POST("/levels", h.CreateLevel)
}

func (h *CatalogHandler) GetMuscleGroups(c *gin.Context) {
	groups, err := database.GetMuscleGroups(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get muscle groups"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(groups))
}

func (h *CatalogHandler) CreateMuscleGroup(c *gin.Context) {
	var req CreateNamedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Name is required"))
		return
	}

	group := &models.MuscleGroup{
		ID:   uuid.New().String(),
		Name: name,
	}
	if err := database.CreateMuscleGroup(h.db, group); err != nil {
		if errors.Is(err, database.ErrMuscleGroupAlreadyExists) {
			c.JSON(http.StatusConflict, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create muscle group"))
		return
	}
	c.JSON(http.StatusCreated, models.NewSuccessResponse(group))
}

func (h *CatalogHandler) GetLevels(c *gin.Context) {
	levels, err := database.GetLevels(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get levels"))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(levels))
}

func (h *CatalogHandler) CreateLevel(c *gin.Context) {
	var req CreateNamedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Name is required"))
		return
	}

	level := &models.Level{
		ID:   uuid.New().String(),
		Name: name,
	}
	if err := database.CreateLevel(h.db, level); err != nil {
		if errors.Is(err, database.ErrLevelAlreadyExists) {
			c.JSON(http.StatusConflict, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create level"))
		return
	}
	c.JSON(http.StatusCreated, models.NewSuccessResponse(level))
}
