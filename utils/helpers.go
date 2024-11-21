package utils

import (
	"fmt"
	"strconv"
	"time"

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
