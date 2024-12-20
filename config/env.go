package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	CDN_URL          string
	AUTH_URL         string
	S3_BUCKET_NAME   string
	S3_ENDPOINT      string
	S3_ACCESS_KEY    string
	S3_SECRET_KEY    string
	S3_IS_HTTPS      bool
	DB_USER          string
	DB_PASS          string
	DB_HOST          string
	DB_PORT          string
	DB_NAME          string
	DB_SSLMODE       string
	STORAGE_PROVIDER string
	SWIFT_USERNAME   string
	SWIFT_API_KEY    string
	SWIFT_AUTH_URL   string
	SWIFT_REGION     string
	SWIFT_CONTAINER  string
}

func LoadEnv() (*Env, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	return &Env{
		CDN_URL:          os.Getenv("CDN_URL"),
		AUTH_URL:         os.Getenv("AUTH_URL"),
		S3_BUCKET_NAME:   os.Getenv("S3_BUCKET_NAME"),
		S3_ENDPOINT:      os.Getenv("S3_ENDPOINT"),
		S3_ACCESS_KEY:    os.Getenv("S3_ACCESS_KEY"),
		S3_SECRET_KEY:    os.Getenv("S3_SECRET_KEY"),
		S3_IS_HTTPS:      os.Getenv("S3_IS_HTTPS") == "true",
		DB_USER:          os.Getenv("DB_USER"),
		DB_PASS:          os.Getenv("DB_PASS"),
		DB_HOST:          os.Getenv("DB_HOST"),
		DB_PORT:          os.Getenv("DB_PORT"),
		DB_NAME:          os.Getenv("DB_NAME"),
		DB_SSLMODE:       os.Getenv("DB_SSLMODE"),
		STORAGE_PROVIDER: os.Getenv("STORAGE_PROVIDER"),
		SWIFT_USERNAME:   os.Getenv("SWIFT_USERNAME"),
		SWIFT_API_KEY:    os.Getenv("SWIFT_API_KEY"),
		SWIFT_AUTH_URL:   os.Getenv("SWIFT_AUTH_URL"),
		SWIFT_REGION:     os.Getenv("SWIFT_REGION"),
		SWIFT_CONTAINER:  os.Getenv("SWIFT_CONTAINER"),
	}, nil
}
