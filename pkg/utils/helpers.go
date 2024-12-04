package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"video-feed/internal/models"

	"github.com/gin-gonic/gin"
)

func GenerateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func GetUserID(c *gin.Context) string {
	// Mock implementation, replace with actual logic
	return "dummy-user-id"
}

func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func ValidateUploadSession(uploadDir string) bool {
	info, err := GetUploadSession(uploadDir)
	if err != nil {
		return false
	}

	// Check if session is not older than 24 hours
	return time.Since(info.CreatedAt) < 24*time.Hour
}

func GetUploadSession(uploadDir string) (*models.ChunkInfo, error) {
	infoFile := fmt.Sprintf("%s/info.json", uploadDir)
	data, err := os.ReadFile(infoFile)
	if err != nil {
		return nil, err
	}

	var info models.ChunkInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func CleanupOldUploads() {
	uploadBaseDir := "tmp/uploads"
	entries, err := os.ReadDir(uploadBaseDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		uploadDir := filepath.Join(uploadBaseDir, entry.Name())
		info, err := GetUploadSession(uploadDir)
		if err != nil {
			// If can't read info, assume it's old and remove it
			os.RemoveAll(uploadDir)
			continue
		}

		// Remove directories older than 24 hours
		if time.Since(info.CreatedAt) > 24*time.Hour {
			os.RemoveAll(uploadDir)
		}
	}
}
