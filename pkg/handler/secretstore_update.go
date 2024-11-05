package handler

import (
	"fmt"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/gin-gonic/gin"
	"sigs.k8s.io/yaml"
)

// SecretStoreUpdateRequest represents the update request with YAML content
type SecretStoreUpdateRequest struct {
	YAML string `json:"yaml" binding:"required"`
}

// UpdateSecretStore handles the update of a SecretStore configuration
func UpdateSecretStore(c *gin.Context) {
	var req SecretStoreUpdateRequest

	// Parse request JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
		return
	}

	// Validate YAML by unmarshaling into SecretStore struct
	secretStore := &esv1beta1.SecretStore{}
	if err := yaml.Unmarshal([]byte(req.YAML), secretStore); err != nil {
		// Try ClusterSecretStore if SecretStore fails
		clusterSecretStore := &esv1beta1.ClusterSecretStore{}
		if err := yaml.Unmarshal([]byte(req.YAML), clusterSecretStore); err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid SecretStore YAML: %v", err)})
			return
		}
	}

	// TODO: write to gitops fs

	// TODO: Implement actual update logic here
	// For now, just return success response
	c.JSON(200, gin.H{
		"message": "SecretStore updated successfully",
		"yaml":    req.YAML,
	})
}
