package services

import (
	"fmt"
	"workouts-backend/src/config"

	"github.com/aranoy15/go-s3"
)

func NewS3Client(cfg *config.Config) (*s3.Client, error) {
	if cfg.S3Config.AccessKeyID == "" || cfg.S3Config.SecretAccessKey == "" {
		return nil, fmt.Errorf("S3 credentials not configured: set S3_ACCESS_KEY_ID and S3_SECRET_ACCESS_KEY")
	}
	if cfg.S3Config.BucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME is required")
	}

	return s3.New(&s3.Config{
		Endpoint:        cfg.S3Config.Endpoint,
		AccessKeyID:     cfg.S3Config.AccessKeyID,
		SecretAccessKey: cfg.S3Config.SecretAccessKey,
		BucketName:      cfg.S3Config.BucketName,
		Region:          cfg.S3Config.Region,
	})
}
