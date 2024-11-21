package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func UploadToMinIO(bucketName, objectName, filePath string) error {
	// Initialize MinIO client
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "password123", ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("Failed to create MinIO client: %v", err)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Upload file to MinIO bucket
	_, err = minioClient.PutObject(
		context.Background(),
		bucketName, // MinIO Bucket Name
		objectName, // Name of the object
		file,       // File to be uploaded
		-1,         // Content size (-1 for unknown)
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		return fmt.Errorf("Failed to upload object: %v", err)
	}

	return nil
}
