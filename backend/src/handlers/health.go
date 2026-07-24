package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHealthHandler(router *gin.RouterGroup) {
	healthHandler := NewHealthHandler()
	router.GET("/health", healthHandler.CheckHealth)
}

type HealthHandler struct {
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) CheckHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
