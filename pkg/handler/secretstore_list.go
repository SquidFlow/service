package handler

import (
	"context"
	"fmt"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
)

type ListSecretStoreResponse struct {
	Success bool                `json:"success"`
	Total   int                 `json:"total"`
	Items   []SecretStoreDetail `json:"items"`
	Message string              `json:"message"`
}

// ListSecretStore returns a list of secret stores
func ListSecretStore(c *gin.Context) {
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

	secretStores, err := RunListSecretStore(context.Background(), &SecretStoreListOptions{
		CloneOpts: cloneOpts,
	})
	if err != nil {
		log.G().Errorf("Failed to list secret stores: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list secret stores: %v", err)})
		return
	}

	c.JSON(200, ListSecretStoreResponse{
		Success: true,
		Total:   len(secretStores),
		Items:   secretStores,
		Message: "secret stores retrieved successfully",
	})
}

type SecretStoreListOptions struct {
	CloneOpts *git.CloneOptions
}

func RunListSecretStore(ctx context.Context, opts *SecretStoreListOptions) ([]SecretStoreDetail, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return nil, err
	}

	matches, err := billyUtils.Glob(repofs, repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		"ss-*.yaml",
	))
	if err != nil {
		return nil, err
	}

	var secretStores []SecretStoreDetail

	for _, file := range matches {
		log.G().WithField("file", file).Debug("Found secret store")

		secretStore := &esv1beta1.SecretStore{}
		if err := repofs.ReadYamls(file, secretStore); err != nil {
			log.G().Warnf("Failed to read secret store from %s: %v", file, err)
			continue
		}

		if secretStore.Kind != "SecretStore" {
			log.G().Warnf("Skip %s: not a SecretStore", file)
			continue
		}

		log.G().WithFields(log.Fields{
			"id":       secretStore.Annotations["squidflow.github.io/id"],
			"name":     secretStore.Name,
			"provider": "vault",
		}).Debug("Found secret store")

		detail := SecretStoreDetail{
			ID:          secretStore.Annotations["squidflow.github.io/id"],
			Name:        secretStore.Name,
			Provider:    "vault",
			Type:        "SecretStore",
			Status:      "Active",
			Path:        *secretStore.Spec.Provider.Vault.Path,
			LastSynced:  secretStore.Annotations["squidflow.github.io/last-synced"],
			CreatedAt:   secretStore.Annotations["squidflow.github.io/created-at"],
			LastUpdated: secretStore.Annotations["squidflow.github.io/updated-at"],
			Environment: []string{"sit", "uat", "prod"},
			Health: SecretStoreHealth{
				Status:  "Healthy",
				Message: "Secret store is operating normally",
			},
		}

		secretStores = append(secretStores, detail)
	}

	return secretStores, nil
}
