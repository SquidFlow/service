package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

type DeleteSecretStoreResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteSecretStore handles the deletion of a SecretStore configuration
func DeleteSecretStore(c *gin.Context) {
	secretStoreID := c.Param("id")
	if secretStoreID == "" {
		c.JSON(400, gin.H{"error": "SecretStore ID is required"})
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

	if err := RunDeleteSecretStore(context.Background(), secretStoreID, &SecretStoreDeleteOptions{
		CloneOpts: cloneOpts,
	}); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete secret store: %v", err)})
		return
	}

	c.JSON(200, DeleteSecretStoreResponse{
		Success: true,
		Message: "secret store deleted successfully",
	})
}

type SecretStoreDeleteOptions struct {
	CloneOpts *git.CloneOptions
}

func RunDeleteSecretStore(ctx context.Context, secretStoreID string, opts *SecretStoreDeleteOptions) error {
	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return err
	}

	secretStorePath := repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		fmt.Sprintf("ss-%s.yaml", secretStoreID),
	)

	exists := repofs.ExistsOrDie(secretStorePath)
	if !exists {
		log.G().Infof("secret store %s not found, considering it as already deleted", secretStoreID)
		return nil
	}

	if err := repofs.Remove(secretStorePath); err != nil {
		return fmt.Errorf("failed to delete secret store file: %v", err)
	}

	if _, err = r.Persist(ctx, &git.PushOptions{
		CommitMsg: fmt.Sprintf("chore: deleted secret store '%s'", secretStoreID),
	}); err != nil {
		return fmt.Errorf("failed to push secret store deletion to repo: %v", err)
	}

	log.G().Infof("secret store deleted: '%s'", secretStoreID)
	return nil
}
