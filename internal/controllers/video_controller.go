package controllers

import (
	"net/http"
	"video-feed/internal/dto"
	"video-feed/internal/services"
	"video-feed/pkg/utils"
	"video-feed/pkg/utils/logger"

	"github.com/gin-gonic/gin"
)

type VideoController struct {
	service *services.VideoService
}

func NewVideoController(service *services.VideoService) *VideoController {
	return &VideoController{service: service}
}

func (vc *VideoController) UploadVideo(c *gin.Context) {
	video, err := vc.service.UploadVideo(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, video)
}

func (vc *VideoController) ListVideo(c *gin.Context) {
	videos, err := vc.service.ListVideo(c)
	if err != nil {
		logger.Log.Error("failed to get videos", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"videos": videos})
}

func (vc *VideoController) InitiateChunkUpload(c *gin.Context) {
	var requestData dto.InitiateChunkDTO
	if err := c.ShouldBindJSON(&requestData); err != nil {
		logger.Log.Error("Invalid data", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data", "err": err.Error()})
		return
	}

	uploadID, err := vc.service.InitiateChunkUpload(requestData)
	if err != nil {
		logger.Log.Error("failed to initiate chunk upload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"uploadId": uploadID})
}

func (vc *VideoController) UploadChunk(c *gin.Context) {
	var dto dto.ChunkUploadDTO
	if err := c.ShouldBind(&dto); err != nil {
		logger.Log.Error("Invalid request format", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err := vc.service.ChunkUpload(dto)
	if err != nil {
		logger.Log.Error("failed to upload chunk", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Chunk uploaded successfully",
		"chunk":  dto.ChunkNumber,
	})
}

func (vc *VideoController) CompleteChunkUpload(c *gin.Context) {
	var requestData dto.CompleteChunkUploadDTO
	if err := c.ShouldBindJSON(&requestData); err != nil {
		logger.Log.Error("Invalid data", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data", "err": err.Error()})
		return
	}

	userId := utils.GetUserID(c)
	video, err := vc.service.CompleteChunkUpload(requestData, userId)
	if err != nil {
		logger.Log.Error("failed to complete chunk upload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error", "err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, video)
}
