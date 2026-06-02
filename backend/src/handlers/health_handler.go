package handlers

import (
	"net/http"

	"workouts-backend/src/database"
	"workouts-backend/src/models"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) CheckHealth(c *gin.Context) {
	sqlDB, err := h.db.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err.Error()))
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.NewErrorResponse("database unavailable"))
		return
	}

	c.JSON(http.StatusOK, models.NewMessageResponse("Service is running"))
}
