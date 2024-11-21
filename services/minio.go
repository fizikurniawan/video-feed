package services

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOService struct {
	Client    *minio.Client
	Bucket    string
	Endpoint  string
	AccessKey string
	SecretKey string
}

func NewMinIOService(endpoint, accessKey, secretKey, bucket string) (*MinIOService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Kalau MinIO-nya nggak pakai HTTPS, pakai false
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	return &MinIOService{
		Client:    client,
		Bucket:    bucket,
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}, nil
}

func (m *MinIOService) UploadObject(objectName string, file *os.File) error {
	ext := filepath.Ext(objectName)
	contentType := mime.TypeByExtension(ext)

	_, err := m.Client.FPutObject(context.Background(), m.Bucket, objectName, file.Name(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}
	return nil
}
