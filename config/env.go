package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	CDN_URL           string
	AUTH_URL          string
	MINIO_BUCKET_NAME string
	MINIO_ENDPOINT    string
	MINIO_ACCESS_KEY  string
	MINIO_SECRET_KEY  string
	MINIO_IS_HTTPS    bool
	DB_USER           string
	DB_PASS           string
	DB_HOST           string
	DB_PORT           string
	DB_NAME           string
	DB_SSLMODE        string
}

func LoadEnv() (*Env, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file")
	}

	return &Env{
		CDN_URL:           os.Getenv("CDN_URL"),
		AUTH_URL:          os.Getenv("AUTH_URL"),
		MINIO_BUCKET_NAME: os.Getenv("MINIO_BUCKET_NAME"),
		MINIO_ENDPOINT:    os.Getenv("MINIO_ENDPOINT"),
		MINIO_ACCESS_KEY:  os.Getenv("MINIO_ACCESS_KEY"),
		MINIO_SECRET_KEY:  os.Getenv("MINIO_SECRET_KEY"),
		MINIO_IS_HTTPS:    os.Getenv("MINIO_IS_HTTPS") == "true",
		DB_USER:           os.Getenv("DB_USER"),
		DB_PASS:           os.Getenv("DB_PASS"),
		DB_HOST:           os.Getenv("DB_HOST"),
		DB_PORT:           os.Getenv("DB_PORT"),
		DB_NAME:           os.Getenv("DB_NAME"),
		DB_SSLMODE:        os.Getenv("DB_SSLMODE"),
	}, nil
}