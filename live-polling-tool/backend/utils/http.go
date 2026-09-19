package utils

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data any) { c.JSON(200, gin.H{"data": data}) }
func Error(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}
