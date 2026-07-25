//go:build wireinject
// +build wireinject

package main

import (
	"workouts-backend/src/config"
	"workouts-backend/src/database"
	"workouts-backend/src/router"
	"workouts-backend/src/services"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitializeApp() (*gin.Engine, error) {
	wire.Build(
		config.Load,
		database.Connect,
		services.NewS3Client,
		router.NewRouter,
	)
	return nil, nil
}
