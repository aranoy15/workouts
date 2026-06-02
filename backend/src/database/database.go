package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"workouts-backend/src/config"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func Connect(cfg *config.Config) (*DB, error) {
	if cfg.DBConfig.Host == "" {
		return nil, fmt.Errorf("database host is not configured")
	}

	schemaName := cfg.DBConfig.Schema
	if schemaName == "" {
		schemaName = "workouts"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.DBName,
		cfg.DBConfig.SSLMode,
		schemaName,
	)

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	setSearchPathSQL := fmt.Sprintf("SET search_path TO %s", schemaName)
	log.Printf("Setting search_path to: %s", schemaName)
	if _, err = sqlDB.ExecContext(ctx, setSearchPathSQL); err != nil {
		return nil, fmt.Errorf("failed to set search_path to %s: %w", schemaName, err)
	}

	db.Callback().Raw().Before("gorm:query").Register("set_search_path", func(db *gorm.DB) {
		setSearchPathSQL := fmt.Sprintf("SET LOCAL search_path TO %s", schemaName)
		db.Exec(setSearchPathSQL)
	})

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected with schema: %s", schemaName)

	return &DB{DB: db}, nil
}
