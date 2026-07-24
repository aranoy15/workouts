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
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	middleware.CORS(r)

	api := r.Group(defaultAPI)
	handlers.RegisterHealthHandler(api)

	log.Printf("Router initialized on port %s", cfg.Port)
	log.Println("  Public endpoints: /api/health")

	return r
}
