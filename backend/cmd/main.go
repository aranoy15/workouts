//go:build !wireinject
// +build !wireinject

package main

import (
	"log"

	"workouts-backend/src/config"
)

func main() {
	cfg := config.Load()

	r, err := InitializeApp()
	if err != nil {
		log.Fatal("Failed to initialize app:", err)
	}

	log.Printf("Server started on port %s", cfg.Port)
	log.Printf("API available at: http://localhost:%s/api", cfg.Port)
	log.Printf("Health check: http://localhost:%s/api/health", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
