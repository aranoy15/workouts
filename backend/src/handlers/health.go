package handlers

import (
	"net/http"
	"workouts-backend/src/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterHealthHandler(router *gin.RouterGroup, db *gorm.DB) {
	healthHandler := NewHealthHandler(db)
	router.GET("/health", healthHandler.CheckHealth)
}

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) CheckHealth(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.NewErrorResponse("database unavailable"))
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.NewErrorResponse("database unavailable"))
		return
	}

	c.JSON(http.StatusOK, models.NewMessageResponse("OK"))
}
