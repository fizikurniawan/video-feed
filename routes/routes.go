package routes

import (
	"net/http"
	"video-feed/config"
	"video-feed/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, cfg *config.AppConfig) {

	videoController := &controllers.VideoController{
		DB:        cfg.DB,
		MinIO:     cfg.MinIO,
		Container: cfg.Container,
		CDNURL:    cfg.CDNURL,
	}

	api := router.Group("/api")
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

	api.GET("/list", videoController.ListVideo)
	api.POST("/upload", videoController.UploadVideo)
	// Chunk upload routes
	api.POST("/initiate-chunk-upload", videoController.InitiateChunkUpload)
	api.POST("/upload-chunk", videoController.UploadChunk)
	api.POST("/complete-chunk-upload", videoController.CompleteChunkUpload)

}
