package storage

import (
	"context"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Service struct {
	Client    *minio.Client
	Bucket    string
	Endpoint  string
	AccessKey string
	SecretKey string
}

func NewS3Service(endpoint, accessKey, secretKey, bucket string, isHTTPS bool) (*S3Service, error) {
	var client *minio.Client
	var err error

	for i := 0; i < 3; i++ {
		client, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: isHTTPS,
		})
		if err == nil {
			break
		}
		log.Printf("Retry %d: Failed to connect to S3: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	// Verify connection
	_, err = client.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}

	return &S3Service{
		Client: client,
		Bucket: bucket,
	}, nil
}

func (m *S3Service) UploadObject(objectName string, file *os.File) error {
	ext := filepath.Ext(objectName)
	contentType := mime.TypeByExtension(ext)

	_, err := m.Client.FPutObject(context.Background(), m.Bucket, objectName, file.Name(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("Failed to upload object %s: %v", objectName, err)
		return err
	}
	return nil
}
