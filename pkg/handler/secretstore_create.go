package handler

import (
	"context"
	"fmt"
	"time"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/util"
)

type SecretStoreCreateReq struct {
	SecretStoreYaml string `json:"secret_store_yaml"`
}

type SecretStoreCreateResponse struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func CreateSecretStore(c *gin.Context) {
	var req SecretStoreCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	want := esv1beta1.SecretStore{}
	err := yaml.Unmarshal([]byte(req.SecretStoreYaml), &want)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to unmarshal SecretStore: %v", err)})
		return
	}

	if want.Spec.Provider == nil {
		c.JSON(400, gin.H{"error": "Provider configuration is required"})
		return
	}

	if want.Spec.Provider.Vault == nil {
		c.JSON(400, gin.H{"error": "Only Vault provider is supported"})
		return
	}

	if want.Annotations != nil && want.Annotations["squidflow.github.io/id"] != "" {
		c.JSON(400, gin.H{"error": "id not allow set via client"})
		return
	}
	if want.Annotations == nil {
		want.Annotations = make(map[string]string)
	}
	want.Annotations["squidflow.github.io/last-synced"] = time.Now().Format(time.RFC3339)
	want.Annotations["squidflow.github.io/created-at"] = time.Now().Format(time.RFC3339)
	want.Annotations["squidflow.github.io/updated-at"] = time.Now().Format(time.RFC3339)
	want.Annotations["squidflow.github.io/id"] = getNewId()

	log.G().WithFields(log.Fields{
		"id": want.Annotations["squidflow.github.io/id"],
	}).Debug("generated id for secret store")

	log.G().WithFields(log.Fields{
		"name":          want.Name,
		"namespace":     want.Namespace,
		"annotations":   want.Annotations,
		"vault_auth":    want.Spec.Provider.Vault.Auth,
		"vault_server":  want.Spec.Provider.Vault.Server,
		"vault_path":    want.Spec.Provider.Vault.Path,
		"vault_version": want.Spec.Provider.Vault.Version,
	}).Debug("Creating SecretStore with Vault provider")

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

	if err := writeSecretStore2Repo(context.Background(), &want, cloneOpts, false); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create external secret: %v", err)})
		return
	}

	c.JSON(201, SecretStoreCreateResponse{
		Name:    want.Name,
		ID:      want.Annotations["squidflow.github.io/id"],
		Success: true,
		Message: "SecretStore created successfully",
	})
}

// create the external secret to gitops repo
func writeSecretStore2Repo(ctx context.Context, ss *esv1beta1.SecretStore, cloneOpts *git.CloneOptions, force bool) error {
	log.G().WithFields(log.Fields{
		"name":      ss.Name,
		"id":        ss.Annotations["squidflow.github.io/id"],
		"cloneOpts": cloneOpts,
		"force":     force,
	}).Debug("clone options")

	r, repofs, err := prepareRepo(ctx, cloneOpts, "")
	if err != nil {
		log.G().WithError(err).Error("failed to prepare repo")
		return err
	}

	ssYaml, err := yaml.Marshal(ss)
	if err != nil {
		log.G().WithError(err).Error("failed to marshal secret store")
		return err
	}

	ssExists := repofs.ExistsOrDie(
		repofs.Join(
			store.Default.BootsrtrapDir,
			store.Default.ClusterResourcesDir,
			store.Default.ClusterContextName,
			fmt.Sprintf("ss-%s.yaml", ss.Annotations["squidflow.github.io/id"]),
		),
	)
	if ssExists && !force {
		return fmt.Errorf("secret store '%s' already exists", ss.GetName())
	}

	bulkWrites := []fs.BulkWriteRequest{}
	bulkWrites = append(bulkWrites, fs.BulkWriteRequest{
		Filename: repofs.Join(
			store.Default.BootsrtrapDir,
			store.Default.ClusterResourcesDir,
			store.Default.ClusterContextName,
			fmt.Sprintf("ss-%s.yaml", ss.Annotations["squidflow.github.io/id"]),
		),
		Data:   util.JoinManifests(ssYaml),
		ErrMsg: "failed to create secret store file",
	})

	if err = fs.BulkWrite(repofs, bulkWrites...); err != nil {
		return err
	}

	if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: fmt.Sprintf("chore: added secret store '%s'", ss.GetName())}); err != nil {
		log.G().WithError(err).Error("failed to push secret store to repo")
		return err
	}

	log.G().Infof("secret store created: '%s'", ss.GetName())

	return nil
}
