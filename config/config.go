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

	rackCfg := storage.RackspaceCfg{
		RACKSPACE_USERNAME:  env.RACKSPACE_USERNAME,
		RACKSPACE_API_KEY:   env.RACKSPACE_API_KEY,
		RACKSPACE_AUTH_URL:  env.RACKSPACE_AUTH_URL,
		RACKSPACE_REGION:    env.RACKSPACE_REGION,
		RACKPSACE_CONTAINER: env.RACKPSACE_CONTAINER,
	}
	miniCfg := storage.MinioCfg{
		MINIO_ENDPOINT:    env.MINIO_ENDPOINT,
		MINIO_ACCESS_KEY:  env.MINIO_ACCESS_KEY,
		MINIO_SECRET_KEY:  env.MINIO_SECRET_KEY,
		MINIO_BUCKET_NAME: env.MINIO_BUCKET_NAME,
		MINIO_IS_HTTPS:    env.MINIO_IS_HTTPS,
	}
	storageSrv, err := storage.NewStorageService(env.STORAGE_PROVIDER, miniCfg, rackCfg)
	if err != nil {
		log.Fatalf("Storage initialization failed: %v", err)
	}

	return &AppConfig{
		DB:      db,
		Env:     env,
		Storage: storageSrv,
	}
}
