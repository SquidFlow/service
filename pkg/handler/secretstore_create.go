package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/h4-poc/service/pkg/log"

	"sigs.k8s.io/yaml"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
)

type SecretStoreCreateReq struct {
	SecretStoreYaml string `json:"secret_store_yaml"`
}

func CreateSecretStore(c *gin.Context) {
	var req SecretStoreCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// SecretStoreYaml
	want := esv1beta1.SecretStore{}
	err := yaml.Unmarshal([]byte(req.SecretStoreYaml), &want)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to unmarshal SecretStore: %v", err)})
		return
	}

	// save the data to gitios under /root application
	log.G().WithFields(log.Fields{
		"secret_store_yaml": req.SecretStoreYaml,
	}).Info("creating SecretStore...")

	c.JSON(201, gin.H{
		"name":    want.Name,
		"message": "SecretStore created successfully",
	})
}

// create the external secret to gitops repo
func createExternalSecret(secretStore *esv1beta1.SecretStore) error {
	return nil
}
