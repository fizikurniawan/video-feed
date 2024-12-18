package routes

import (
	"net/http"
	"video-feed/config"
	"video-feed/internal/controllers"
	"video-feed/internal/repositories"
	"video-feed/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, cfg *config.AppConfig) {
	videoRepo := repositories.NewVideoRepository(cfg.DB)
	videoService := services.NewVideoService(videoRepo, cfg.Storage, cfg)
	videoController := controllers.NewVideoController(videoService)

	api := router.Group("/api")
	api.POST("/upload", videoController.UploadVideo)
	api.GET("/list", videoController.ListVideo)
	api.POST("/initiate-chunk-upload", videoController.InitiateChunkUpload)
	api.POST("/upload-chunk", videoController.UploadChunk)
	api.POST("/complete-chunk-upload", videoController.CompleteChunkUpload)

	// api.Use(middlewares.AuthMiddleware())

	web := router.Group("/web")

	// Route untuk serve HTML
	web.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "feed.html", nil)
	})
	// Route untuk serve HTML
	web.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", nil)
	})
}
