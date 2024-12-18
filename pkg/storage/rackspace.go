package storage

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/ncw/swift/v2"
)

type RackspaceService struct {
	Client    *swift.Connection
	Username  string
	ApiKey    string
	AuthURL   string
	Region    string
	Container string
}

func NewRackspaceService(username, apiKey, authURL, region, containerName string) (*RackspaceService, error) {
	// Input validation
	if username == "" || apiKey == "" || authURL == "" || region == "" {
		return nil, fmt.Errorf("missing required authentication parameters")
	}

	// Create connection with retry mechanism
	var client *swift.Connection
	var err error

	for attempts := 0; attempts < 3; attempts++ {
		client = &swift.Connection{
			UserName: username,
			ApiKey:   apiKey,
			AuthUrl:  authURL,
			Region:   region,
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt authentication
		err = client.Authenticate(ctx)
		if err == nil {
			break
		}

		log.Printf("Authentication attempt %d failed: %v", attempts+1, err)

		// Exponential backoff
		time.Sleep(time.Duration(attempts+1) * 2 * time.Second)
	}

	// Return error if all attempts fail
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate after 3 attempts: %w", err)
	}

	// Verify connection by listing containers
	ctx := context.Background()
	_, err = client.ContainerNames(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	return &RackspaceService{
		Client:    client,
		Username:  username,
		ApiKey:    apiKey,
		AuthURL:   authURL,
		Region:    region,
		Container: containerName,
	}, nil
}

func (r *RackspaceService) UploadObject(objectName string, file *os.File) error {
	ext := filepath.Ext(objectName)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Upload object
	_, err := r.Client.ObjectPut(ctx, r.Container, objectName, file, false,
		"", contentType, swift.Headers{
			"X-Object-Meta-Uploaded-By": r.Username,
			"X-Object-Meta-Upload-Date": time.Now().UTC().Format(time.RFC3339),
		})
	if err != nil {
		return fmt.Errorf("failed to upload object %s to %s: %w", objectName, r.Container, err)
	}

	log.Printf("Successfully uploaded object %s to container %s", objectName, r.Container)
	return nil
}
