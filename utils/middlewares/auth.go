package middlewares

import (
	"net/http"
	"video-feed/config"

	"github.com/gin-gonic/gin"
)

func isTokenValid(authHeader string, authUrl string) (bool, error) {
	client := http.Client{}

	// Buat request
	req, err := http.NewRequest("GET", authUrl, nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", authHeader)
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()
	return res.StatusCode == 200, nil

}

func AuthMiddleware(cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		isValid, _ := isTokenValid(authHeader, cfg.AUTHURL)
		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
