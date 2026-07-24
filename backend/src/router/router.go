package router

import (
	"log"

	"workouts-backend/src/config"
	"workouts-backend/src/database"
	"workouts-backend/src/handlers"
	"workouts-backend/src/middleware"
	"workouts-backend/src/models"

	"github.com/gin-gonic/gin"
)

const defaultAPI = "/api"

func NewRouter(
	cfg *config.Config,
	db *database.DB,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	middleware.CORS(r)

	api := r.Group(defaultAPI)
	handlers.RegisterHealthHandler(api, db.DB)
	handlers.RegisterUserHandler(api, cfg, db.DB)

	admin := api.Group("")
	middleware.Auth(admin, cfg, db.DB, string(models.UserRoleAdmin))
	handlers.RegisterAdminUserHandler(admin, cfg, db.DB)

	log.Printf("Router initialized on port %s", cfg.Port)
	log.Println("  Public endpoints: /api/health, /api/auth/login")
	log.Println("  Admin endpoints: /api/users")

	return r
}
