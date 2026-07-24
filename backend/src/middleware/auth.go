package middleware

import (
	"net/http"
	"slices"
	"strings"
	"workouts-backend/src/config"
	"workouts-backend/src/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer"
)

func Auth(router gin.IRouter, cfg *config.Config, db *gorm.DB, allowedRoles ...string) {
	router.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		if !strings.HasPrefix(token, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		token = strings.TrimPrefix(token, bearerPrefix+" ")
		claims, err := jwt.Parse(token, func(taken *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		if !claims.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := claims.Claims.(jwt.MapClaims)["user_id"].(string)
		user, err := database.GetUserByID(db, userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if len(allowedRoles) > 0 && !slices.Contains(allowedRoles, string(user.Role)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Set("user", user)
		c.Next()
	})
}
