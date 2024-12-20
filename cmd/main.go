package main

import (
	"time"
	"video-feed/config"
	"video-feed/pkg/utils/logger"
	"video-feed/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configurations
	appConfig := config.LoadConfig()

	// Setup Gin
	router := gin.Default()
	router.Use(gin.Recovery())

	router.LoadHTMLGlob("templates/*")

	// Setup CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow semua origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour, // Cache preflight response
	}))

	// Register routes
	routes.RegisterRoutes(router, appConfig)

	// Start server
	logger.Log.Info("Starting the application...")
	router.Run(":8080")
}
