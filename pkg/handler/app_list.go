package handler

import (
	"github.com/gin-gonic/gin"
)

func ListApplications(c *gin.Context) {
	c.JSON(200, gin.H{"applications": []Application{}})
}
