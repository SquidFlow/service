package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListAppCode handles the request to list app codes
func ListAppCode(c *gin.Context) {
	appCodes := []string{"esfs", "esfs-dev", "esfs-test"}
	c.JSON(http.StatusOK, gin.H{
		"appCodes": appCodes,
	})
	return
}
