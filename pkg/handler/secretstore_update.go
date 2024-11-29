package handler

import (
	"context"
	"fmt"
	"time"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/store"
)

type SecretStoreUpdateRequest struct {
	Name    string                        `json:"name,omitempty"`
	Path    string                        `json:"path,omitempty"`
	Auth    *esv1beta1.VaultAuth          `json:"auth,omitempty"`
	Server  string                        `json:"server,omitempty"`
	Version esv1beta1.VaultKVStoreVersion `json:"version,omitempty"`
}

type SecretStoreUpdateResponse struct {
	Item    SecretStoreDetail `json:"item"`
	Success bool              `json:"success"`
	Message string            `json:"message"`
}

func UpdateSecretStore(c *gin.Context) {
	secretStoreID := c.Param("id")
	if secretStoreID == "" {
		c.JSON(400, gin.H{"error": "secret store ID is required"})
		return
	}

	var req SecretStoreUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: true,
	}
	cloneOpts.Parse()

	_, repofs, err := prepareRepo(context.Background(), cloneOpts, "")
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to prepare repo: %v", err)})
		return
	}

	secretStorePath := repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		fmt.Sprintf("ss-%s.yaml", secretStoreID),
	)

	secretStore := &esv1beta1.SecretStore{}
	if err := repofs.ReadYamls(secretStorePath, secretStore); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to read secret store: %v", err)})
		return
	}

	if req.Name != "" {
		secretStore.Name = req.Name
	}
	if req.Path != "" {
		secretStore.Spec.Provider.Vault.Path = &req.Path
	}
	if req.Auth != nil {
		secretStore.Spec.Provider.Vault.Auth = *req.Auth
	}
	if req.Server != "" {
		secretStore.Spec.Provider.Vault.Server = req.Server
	}

	secretStore.Annotations["squidflow.github.io/updated-at"] = time.Now().Format(time.RFC3339)

	if err := writeSecretStore2Repo(context.Background(), secretStore, cloneOpts, true); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to write secret store to repo: %v", err)})
		return
	}

	c.JSON(200, SecretStoreUpdateResponse{
		Item: SecretStoreDetail{
			ID:          secretStore.Annotations["squidflow.github.io/id"],
			Name:        secretStore.Name,
			Provider:    "vault",
			Type:        "SecretStore",

			Status:      "Active",
			Path:        *secretStore.Spec.Provider.Vault.Path,
			LastSynced:  secretStore.Annotations["squidflow.github.io/last-synced"],
			CreatedAt:   secretStore.Annotations["squidflow.github.io/created-at"],
			LastUpdated: secretStore.Annotations["squidflow.github.io/updated-at"],
			Health: SecretStoreHealth{
				Status:  "Healthy",
				Message: "Secret store updated successfully",
			},
		},
		Success: true,
		Message: "secret store updated successfully",
	})
}
