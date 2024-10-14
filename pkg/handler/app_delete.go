package handler

import (
	"github.com/gin-gonic/gin"
)

func DeleteApplication(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Application deleted"})
}
