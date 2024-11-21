package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"video-feed/services"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type AppConfig struct {
	DB        *sql.DB
	MinIO     *services.MinIOService
	Container string
	CDNURL    string
	AUTHURL   string
}

func LoadConfig() *AppConfig {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Construct DB connection string
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, dbSSLMode,
	)

	// Setup database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// MinIO setup
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")

	minioService, err := services.NewMinIOService(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO service: %v", err)
	}

	return &AppConfig{
		DB:        db,
		MinIO:     minioService,
		Container: minioBucket,
		CDNURL:    os.Getenv("CDN_URL"),
		AUTHURL:   os.Getenv("AUTH_URL"),
	}
}
