package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"time"
	"video-feed/config"
	"video-feed/internal/dto"
	"video-feed/internal/models"
	"video-feed/internal/repositories"
	"video-feed/pkg/storage"
	"video-feed/pkg/utils"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

type VideoService struct {
	repo    *repositories.VideoRepository
	storage storage.StorageService
	cfg     *config.AppConfig
}

func NewVideoService(repo *repositories.VideoRepository, storage storage.StorageService, cfg *config.AppConfig) *VideoService {
	return &VideoService{
		repo:    repo,
		storage: storage,
		cfg:     cfg,
	}
}

func (vs *VideoService) UploadVideo(c *gin.Context) (models.Video, error) {
	file, err := c.FormFile("video")
	if err != nil {
		return models.Video{}, fmt.Errorf("failed to get video: %v", err)
	}

	srcFile, err := file.Open()
	if err != nil {
		return models.Video{}, fmt.Errorf("failed to open video: %v", err)
	}
	defer srcFile.Close()

	mtype, err := mimetype.DetectReader(srcFile)
	if err != nil {
		return models.Video{}, fmt.Errorf("failed to get extension video: %v", err)
	}

	videoID := utils.GenerateUniqueID()
	originalPath := "videos/" + videoID + "/original" + mtype.Extension()

	tmpFilePath := "tmp/" + videoID + mtype.Extension()
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return models.Video{}, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tmpFile.Close()

	_, err = srcFile.Seek(0, 0)
	if err != nil {
		return models.Video{}, fmt.Errorf("failed to reset file pointer: %v", err)
	}
	_, err = tmpFile.ReadFrom(srcFile)
	if err != nil {
		return models.Video{}, fmt.Errorf("failed to write video to temp file: %v", err)
	}

	err = vs.storage.UploadObject(originalPath, tmpFile)
	if err != nil {
		return models.Video{}, fmt.Errorf("failed to upload video to S3: %v", err)
	}

	video := models.Video{
		ID:           videoID,
		UserID:       utils.GetUserID(c),
		OriginalURL:  vs.cfg.Env.CDN_URL + originalPath,
		HLSURL:       vs.cfg.Env.CDN_URL + "videos/" + videoID + "/playlist.m3u8",
		CreatedAt:    time.Now(),
		Description:  c.PostForm("description"),
		Qualities:    []string{"original"},
		HLSProcessed: false,
	}

	if err := vs.repo.Create(&video); err != nil {
		return models.Video{}, fmt.Errorf("failed to save video metadata: %v", err)
	}

	return video, nil
}

func (vs *VideoService) ListVideo(c *gin.Context) ([]models.Video, error) {
	videos, err := vs.repo.ListUserVideos(utils.GetUserID(c), 100, 0)
	return videos, err
}

func (vs *VideoService) InitiateChunkUpload(dto dto.InitiateChunkDTO) (string, error) {
	uploadID := utils.GenerateUniqueID()
	fileName := dto.FileName
	totalChunks := dto.TotalChunks

	println("anjeng", fileName, uploadID, totalChunks)

	// Create directory for this upload
	uploadDir := fmt.Sprintf("tmp/uploads/%s", uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Save upload info to JSON file
	sessionInfo := models.ChunkInfo{
		UploadID:    uploadID,
		TotalChunks: totalChunks,
		FileName:    fileName,
		CreatedAt:   time.Now(),
	}

	infoFile := fmt.Sprintf("%s/info.json", uploadDir)
	infoData, err := json.Marshal(sessionInfo)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(infoFile, infoData, 0644); err != nil {
		return "", err
	}

	return uploadID, nil
}

func (vs *VideoService) SaveChunk(file *multipart.FileHeader, chunkPath string) error {
	// Simpan file chunk ke lokasi yang sudah ditentukan
	dst, err := os.Create(chunkPath)
	if err != nil {
		return fmt.Errorf("failed to create chunk file: %v", err)
	}
	defer dst.Close()

	// Buka file yang di-upload
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open chunk file: %v", err)
	}
	defer src.Close()

	// Salin file ke destination
	_, err = dst.ReadFrom(src)
	if err != nil {
		return fmt.Errorf("failed to save chunk file: %v", err)
	}

	return nil
}

func (vs *VideoService) ChunkUpload(dto dto.ChunkUploadDTO) error {
	// Validate upload session
	uploadDir := fmt.Sprintf("tmp/uploads/%s", dto.UploadID)
	if !utils.ValidateUploadSession(uploadDir) {
		return errors.New("invalid or expired upload session")
	}

	// Save chunk file
	chunkPath := fmt.Sprintf("%s/chunk_%s", uploadDir, dto.ChunkNumber)
	if err := vs.SaveChunk(dto.Chunk, chunkPath); err != nil {
		return err
	}
	return nil
}

func (vs *VideoService) CompleteChunkUpload(dto dto.CompleteChunkUploadDTO, userId string) (*models.Video, error) {
	uploadID := dto.UploadID
	uploadDir := fmt.Sprintf("tmp/uploads/%s", uploadID)

	// Validate upload session
	sessionInfo, err := utils.GetUploadSession(uploadDir)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired upload session")
	}

	// Verify all chunks are present
	for i := 0; i < sessionInfo.TotalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", uploadDir, i)
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("missing chunk %d", i)
		}
	}

	// Combine chunks
	finalPath := fmt.Sprintf("tmp/%s_%s", uploadID, sessionInfo.FileName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create final file")
	}
	defer finalFile.Close()

	// Combine all chunks in order
	for i := 0; i < sessionInfo.TotalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", uploadDir, i)
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk")
		}
		if _, err := finalFile.Write(chunkData); err != nil {
			return nil, fmt.Errorf("failed to write chunk to final file")
		}
	}

	// Get file type
	finalFile.Seek(0, 0)
	mtype, err := mimetype.DetectReader(finalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to detect file type")
	}

	// Generate video ID and paths
	videoID := utils.GenerateUniqueID()
	originalPath := "videos/" + videoID + "/original" + mtype.Extension()

	// Upload to S3
	finalFile.Seek(0, 0)
	err = vs.storage.UploadObject(originalPath, finalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to storage: %v", err)
	}

	// Create video record
	cdnUrl := vs.cfg.Env.CDN_URL
	video := models.Video{
		ID:           videoID,
		UserID:       userId,
		OriginalURL:  cdnUrl + originalPath,
		HLSURL:       "",
		CreatedAt:    time.Now(),
		Description:  dto.Description,
		Qualities:    []string{"original"},
		HLSProcessed: false,
	}

	if err := vs.repo.Create(&video); err != nil {
		return nil, fmt.Errorf("failed to save video metadata")
	}

	// Start HLS processing in background
	hlsJob := NewHLSBackgroundJob(vs.cfg, vs.storage, vs.repo)

	go func() {
		defer os.RemoveAll(uploadDir) // Clean up chunks
		defer os.Remove(finalPath)    // Clean up final file

		resultChan := hlsJob.ProcessHLSWithTimeout(videoID, finalPath)
		result := <-resultChan

		err := hlsJob.HandleJobResult(result)
		if err != nil {
			println(err.Error())
		}
	}()
	return &video, nil
}
