package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"video-feed/models"
	"video-feed/services"
	"video-feed/utils"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

type VideoController struct {
	DB        *sql.DB
	MinIO     *services.MinIOService
	Container string
	CDNURL    string
}

func (vc *VideoController) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get video"})
		return
	}

	srcFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open video"})
		return
	}
	defer srcFile.Close()

	// check mimetype
	mtype, err := mimetype.DetectReader(srcFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get extension video"})
		return
	}

	videoID := utils.GenerateUniqueID()
	originalPath := "videos/" + videoID + "/original" + mtype.Extension()

	// Save file locally first
	tmpFilePath := "tmp/" + videoID + mtype.Extension()
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
		return
	}
	defer tmpFile.Close()

	// Copy file content to temp file
	_, err = srcFile.Seek(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file pointer"})
		return
	}
	_, err = tmpFile.ReadFrom(srcFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write video to temp file"})
		return
	}

	// Upload video to MinIO
	err = vc.MinIO.UploadObject(originalPath, tmpFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload video to MinIO"})
		return
	}

	video := models.Video{
		ID:           videoID,
		UserID:       utils.GetUserID(c),
		OriginalURL:  vc.CDNURL + "/" + originalPath,
		HLSURL:       vc.CDNURL + "/videos/" + videoID + "/playlist.m3u8",
		CreatedAt:    time.Now(),
		Description:  c.PostForm("description"),
		Qualities:    []string{"original"},
		HLSProcessed: false,
	}
	if err := models.SaveVideo(vc.DB, &video); err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata"})
		return
	}

	hlsJob := &services.HLSBackgroundJob{
		MinIO: vc.MinIO,
		DB:    vc.DB,
	}

	go func() {
		defer os.Remove(tmpFilePath)

		resultChan := hlsJob.ProcessHLSWithTimeout(videoID, tmpFilePath)
		result := <-resultChan

		hlsJob.HandleJobResult(result)
	}()

	c.JSON(http.StatusOK, video)
}

func (vc *VideoController) ListVideo(c *gin.Context) {
	videos, err := models.ListUserVideos(vc.DB, utils.GetUserID(c), 100, 0)

	if err != nil {
		println("error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"videos": videos})
}

// ChunkInfo stores information about a chunk upload session
type ChunkInfo struct {
	UploadID    string    `json:"uploadId"`
	ChunkNumber int       `json:"chunkNumber"`
	TotalChunks int       `json:"totalChunks"`
	FileName    string    `json:"fileName"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (vc *VideoController) InitiateChunkUpload(c *gin.Context) {

	var requestData struct {
		FileName    string `json:"fileName"`
		TotalChunks int    `json:"totalChunks"`
	}

	// Bind JSON body to requestData struct
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data", "err": err.Error()})
		return
	}

	uploadID := utils.GenerateUniqueID()
	fileName := requestData.FileName
	totalChunks := requestData.TotalChunks

	println("anjeng", fileName, uploadID, totalChunks)

	// Create directory for this upload
	uploadDir := fmt.Sprintf("tmp/uploads/%s", uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save upload info to JSON file
	sessionInfo := ChunkInfo{
		UploadID:    uploadID,
		TotalChunks: totalChunks,
		FileName:    fileName,
		CreatedAt:   time.Now(),
	}

	infoFile := fmt.Sprintf("%s/info.json", uploadDir)
	infoData, err := json.Marshal(sessionInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session info"})
		return
	}

	if err := os.WriteFile(infoFile, infoData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session info"})
		return
	}

	// Start cleanup goroutine for old uploads (older than 24h)
	go vc.cleanupOldUploads()

	c.JSON(http.StatusOK, gin.H{"uploadId": uploadID})
}

func (vc *VideoController) UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	chunkNumber := utils.StringToInt(c.PostForm("chunkNumber"))

	// Validate upload session
	uploadDir := fmt.Sprintf("tmp/uploads/%s", uploadID)
	if !vc.validateUploadSession(uploadDir) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired upload session"})
		return
	}

	// Get the chunk file
	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get chunk file"})
		return
	}

	// Save chunk file
	chunkPath := fmt.Sprintf("%s/chunk_%d", uploadDir, chunkNumber)
	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chunk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Chunk uploaded successfully",
		"chunk":  chunkNumber,
	})
}

func (vc *VideoController) CompleteChunkUpload(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	uploadDir := fmt.Sprintf("tmp/uploads/%s", uploadID)

	// Validate upload session
	sessionInfo, err := vc.getUploadSession(uploadDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired upload session"})
		return
	}

	// Verify all chunks are present
	for i := 0; i < sessionInfo.TotalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", uploadDir, i)
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Missing chunk %d", i)})
			return
		}
	}

	// Combine chunks
	finalPath := fmt.Sprintf("tmp/%s_%s", uploadID, sessionInfo.FileName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create final file"})
		return
	}
	defer finalFile.Close()

	// Combine all chunks in order
	for i := 0; i < sessionInfo.TotalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", uploadDir, i)
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read chunk"})
			return
		}
		if _, err := finalFile.Write(chunkData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write chunk to final file"})
			return
		}
	}

	// Get file type
	finalFile.Seek(0, 0)
	mtype, err := mimetype.DetectReader(finalFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to detect file type"})
		return
	}

	// Generate video ID and paths
	videoID := utils.GenerateUniqueID()
	originalPath := "videos/" + videoID + "/original" + mtype.Extension()

	// Upload to MinIO
	finalFile.Seek(0, 0)
	err = vc.MinIO.UploadObject(originalPath, finalFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to storage"})
		return
	}

	// Create video record
	video := models.Video{
		ID:           videoID,
		UserID:       utils.GetUserID(c),
		OriginalURL:  vc.CDNURL + "/" + originalPath,
		HLSURL:       vc.CDNURL + "/videos/" + videoID + "/playlist.m3u8",
		CreatedAt:    time.Now(),
		Description:  c.PostForm("description"),
		Qualities:    []string{"original"},
		HLSProcessed: false,
	}

	if err := models.SaveVideo(vc.DB, &video); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata"})
		return
	}

	// Start HLS processing in background
	hlsJob := &services.HLSBackgroundJob{
		MinIO: vc.MinIO,
		DB:    vc.DB,
	}

	go func() {
		defer os.RemoveAll(uploadDir) // Clean up chunks
		defer os.Remove(finalPath)    // Clean up final file

		resultChan := hlsJob.ProcessHLSWithTimeout(videoID, finalPath)
		result := <-resultChan

		hlsJob.HandleJobResult(result)
	}()

	c.JSON(http.StatusOK, video)
}

// Helper functions

func (vc *VideoController) validateUploadSession(uploadDir string) bool {
	info, err := vc.getUploadSession(uploadDir)
	if err != nil {
		return false
	}

	// Check if session is not older than 24 hours
	return time.Since(info.CreatedAt) < 24*time.Hour
}

func (vc *VideoController) getUploadSession(uploadDir string) (*ChunkInfo, error) {
	infoFile := fmt.Sprintf("%s/info.json", uploadDir)
	data, err := os.ReadFile(infoFile)
	if err != nil {
		return nil, err
	}

	var info ChunkInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (vc *VideoController) cleanupOldUploads() {
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
		info, err := vc.getUploadSession(uploadDir)
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
