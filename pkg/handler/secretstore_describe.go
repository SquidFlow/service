package handler

import (
	"context"
	"fmt"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

type DescribeSecretStoreResponse struct {
	Item    SecretStoreDetail `json:"item"`
	Success bool              `json:"success"`
	Message string            `json:"message"`
}

type SecretStoreHealth struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SecretStoreDetail struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Provider    string            `json:"provider"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Environment []string          `json:"environment"`
	Path        string            `json:"path,omitempty"`
	LastSynced  string            `json:"lastSynced"`
	CreatedAt   string            `json:"createdAt"`
	LastUpdated string            `json:"lastUpdated"`
	Health      SecretStoreHealth `json:"health"`
}

func DescribeSecretStore(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "SecretStore ID is required"})
		return
	}

	log.G().WithField("id", id).Debug("describe secret store")

	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: false,
	}
	cloneOpts.Parse()

	secretStore, err := GetSecretStoreFromRepo(context.Background(), &SecretStoreGetOptions{
		CloneOpts: cloneOpts,
		ID:        id,
	})
	if err != nil {
		log.G().Errorf("Failed to get secret store: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get secret store: %v", err)})
		return
	}

	if secretStore == nil {
		c.JSON(404, gin.H{"error": "secret store not found"})
		return
	}

	c.JSON(200, DescribeSecretStoreResponse{
		Success: true,
		Item:    *secretStore,
		Message: "secret store retrieved successfully",
	})
}

type SecretStoreGetOptions struct {
	CloneOpts *git.CloneOptions
	ID        string
}

func GetSecretStoreFromRepo(ctx context.Context, opts *SecretStoreGetOptions) (*SecretStoreDetail, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return nil, err
	}

	secretStorePath := repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		fmt.Sprintf("ss-%s.yaml", opts.ID),
	)

	secretStore := &esv1beta1.SecretStore{}
	if err := repofs.ReadYamls(secretStorePath, secretStore); err != nil {
		return nil, fmt.Errorf("failed to read secret store %s: %v", opts.ID, err)
	}

	if secretStore.Kind != "SecretStore" {
		return nil, fmt.Errorf("invalid secret store kind: %s", secretStore.Kind)
	}

	return &SecretStoreDetail{
		ID:          secretStore.Annotations["h4-poc.github.io/id"],
		Name:        secretStore.Name,
		Provider:    "vault",
		Status:      "Active",
		Path:        *secretStore.Spec.Provider.Vault.Path,
		Type:        "SecretStore",
		Environment: []string{"sit", "uat", "prod"},
		LastSynced:  secretStore.Annotations["h4-poc.github.io/last-synced"],
		CreatedAt:   secretStore.Annotations["h4-poc.github.io/created-at"],
		LastUpdated: secretStore.Annotations["h4-poc.github.io/updated-at"],
		Health: SecretStoreHealth{
			Status:  "Healthy", // 可以根据实际状态判断
			Message: "Secret store is operating normally",
		},
	}, nil
}
