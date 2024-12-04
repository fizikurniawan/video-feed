package config

import (
	"fmt"
	"log"
	"time"
	"video-feed/pkg/database"
	"video-feed/pkg/storage"

	_ "github.com/lib/pq"
)

type AppConfig struct {
	DB    *database.DatabaseManager
	MinIO *storage.MinIOService
	Env   *Env
}

func LoadConfig() *AppConfig {
	// Load environment variables
	env, err := LoadEnv()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env.DB_USER, env.DB_PASS, env.DB_HOST, env.DB_PORT, env.DB_NAME, env.DB_SSLMODE,
	)

	db, err := database.NewDatabaseManager(dbConnString)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	var minioService *storage.MinIOService
	for i := 0; i < 3; i++ {
		minioService, err = storage.NewMinIOService(env.MINIO_ENDPOINT, env.MINIO_ACCESS_KEY, env.MINIO_SECRET_KEY, env.MINIO_BUCKET_NAME, env.MINIO_IS_HTTPS)
		if err == nil {
			break
		}
		log.Printf("Retry %d: Failed to connect to MinIO: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to initialize MinIO service after retries: %v", err)
	}

	return &AppConfig{
		DB:    db,
		MinIO: minioService,
		Env:   env,
	}
}
