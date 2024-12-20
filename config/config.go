package config

import (
	"fmt"
	"log"
	"video-feed/pkg/database"
	"video-feed/pkg/storage"

	_ "github.com/lib/pq"
)

type AppConfig struct {
	DB      *database.DatabaseManager
	Storage storage.StorageService
	Env     *Env
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

	swiftCfg := storage.SwiftCfg{
		SWIFT_USERNAME:  env.SWIFT_USERNAME,
		SWIFT_API_KEY:   env.SWIFT_API_KEY,
		SWIFT_AUTH_URL:  env.SWIFT_AUTH_URL,
		SWIFT_REGION:    env.SWIFT_REGION,
		SWIFT_CONTAINER: env.SWIFT_CONTAINER,
	}
	miniCfg := storage.S3Cfg{
		S3_ENDPOINT:    env.S3_ENDPOINT,
		S3_ACCESS_KEY:  env.S3_ACCESS_KEY,
		S3_SECRET_KEY:  env.S3_SECRET_KEY,
		S3_BUCKET_NAME: env.S3_BUCKET_NAME,
		S3_IS_HTTPS:    env.S3_IS_HTTPS,
	}
	storageSrv, err := storage.NewStorageService(env.STORAGE_PROVIDER, miniCfg, swiftCfg)
	if err != nil {
		log.Fatalf("Storage initialization failed: %v", err)
	}

	return &AppConfig{
		DB:      db,
		Env:     env,
		Storage: storageSrv,
	}
}
