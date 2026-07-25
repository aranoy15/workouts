package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const insecureDefaultJWTSecret = "no-key"

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Schema   string
}

type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
}

type Config struct {
	Port      string
	JWTSecret string
	DBConfig  DBConfig
	S3Config  S3Config
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env not found, using environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" || jwtSecret == insecureDefaultJWTSecret {
		return nil, fmt.Errorf("JWT_SECRET must be set to a non-default value")
	}

	return &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: jwtSecret,
		DBConfig: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "workouts"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Schema:   getEnv("DB_SCHEMA", "workouts"),
		},
		S3Config: S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", "https://storage.yandexcloud.net"),
			AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
			BucketName:      os.Getenv("S3_BUCKET_NAME"),
			Region:          getEnv("S3_REGION", "ru-central1"),
		},
	}, nil
}

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
