package handler

import (
	"github.com/gin-gonic/gin"
)

// DeleteSecretStore handles the update of a SecretStore configuration
func DeleteSecretStore(c *gin.Context) {
	secretStoreName := c.Query("name")

	if secretStoreName == "" {
		c.JSON(400, gin.H{"error": "SecretStore name is required"})
		return
	}
}
