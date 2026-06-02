package router

import (
	"log"

	"workouts-backend/src/config"
	"workouts-backend/src/handlers"
	"workouts-backend/src/middleware"

	"github.com/gin-gonic/gin"
)

const defaultAPI = "/api"

func NewRouter(
	cfg *config.Config,
	healthHandler *handlers.HealthHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	middleware.CORS(r)

	api := r.Group(defaultAPI)
	api.GET("/health", healthHandler.CheckHealth)

	log.Printf("Router initialized on port %s", cfg.Port)
	log.Println("  Public endpoints: /api/health")

	return r
}
