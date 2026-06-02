//go:build wireinject
// +build wireinject

package main

import (
	"workouts-backend/src/config"
	"workouts-backend/src/database"
	"workouts-backend/src/handlers"
	"workouts-backend/src/router"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitializeApp() (*gin.Engine, error) {
	wire.Build(
		config.Load,
		database.Connect,
		handlers.NewHealthHandler,
		router.NewRouter,
	)
	return nil, nil
}
