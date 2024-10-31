package handler

import (
	"github.com/gin-gonic/gin"
)

// TODO: CreateHelmTemplate godoc
func CreateHelmTemplate(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Helm template created"})
	return
}
