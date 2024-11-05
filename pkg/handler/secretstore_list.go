package handler

import (
	"github.com/gin-gonic/gin"
)

// SecretStoreHealth represents the health status of a secret store
type SecretStoreHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// SecretStore represents a secret store configuration
type SecretStore struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Provider    string            `json:"provider"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Path        string            `json:"path,omitempty"`
	LastSynced  string            `json:"lastSynced"`
	CreatedAt   string            `json:"createdAt"`
	LastUpdated string            `json:"lastUpdated"`
	Health      SecretStoreHealth `json:"health"`
}

// ListSecretStore returns a list of secret stores
func ListSecretStore(c *gin.Context) {
	mockData := []SecretStore{
		{
			ID:          1,
			Name:        "aws-secrets-manager",
			Provider:    "AWS",
			Type:        "SecretStore",
			Status:      "Active",
			Path:        "aws/data/applications/*",
			LastSynced:  "2024-03-15T08:30:00Z",
			CreatedAt:   "2024-01-15",
			LastUpdated: "2024-03-15",
			Health: SecretStoreHealth{
				Status:  "Healthy",
				Message: "Connected to AWS Secrets Manager",
			},
		},
		{
			ID:          2,
			Name:        "vault-kv",
			Provider:    "Vault",
			Type:        "ClusterSecretStore",
			Status:      "Active",
			Path:        "secret/data/applications",
			LastSynced:  "2024-03-15T08:25:00Z",
			CreatedAt:   "2024-01-20",
			LastUpdated: "2024-03-15",
			Health: SecretStoreHealth{
				Status:  "Healthy",
				Message: "Connected to Vault server",
			},
		},
		{
			ID:          3,
			Name:        "azure-key-vault",
			Provider:    "Azure",
			Type:        "SecretStore",
			Status:      "Active",
			Path:        "azure/data/certificates",
			LastSynced:  "2024-03-15T08:20:00Z",
			CreatedAt:   "2024-02-01",
			LastUpdated: "2024-03-15",
			Health: SecretStoreHealth{
				Status:  "Warning",
				Message: "High latency detected",
			},
		},
		{
			ID:          4,
			Name:        "gcp-secret-manager",
			Provider:    "GCP",
			Type:        "ClusterSecretStore",
			Status:      "Active",
			Path:        "gcp/data/projects/*/secrets",
			LastSynced:  "2024-03-15T08:15:00Z",
			CreatedAt:   "2024-02-15",
			LastUpdated: "2024-03-15",
			Health: SecretStoreHealth{
				Status:  "Healthy",
				Message: "Connected to GCP Secret Manager",
			},
		},
		{
			ID:          5,
			Name:        "cyberark-conjur",
			Provider:    "CyberArk",
			Type:        "SecretStore",
			Status:      "Error",
			Path:        "cyberark/data/apps/credentials",
			LastSynced:  "2024-03-15T07:00:00Z",
			CreatedAt:   "2024-03-01",
			LastUpdated: "2024-03-15",
			Health: SecretStoreHealth{
				Status:  "Error",
				Message: "Authentication failed",
			},
		},
	}

	c.JSON(200, mockData)
}
