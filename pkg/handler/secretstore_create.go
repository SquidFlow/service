package handler

import (
	"context"
	"fmt"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
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

	if want.Spec.Provider == nil {
		c.JSON(400, gin.H{"error": "Provider configuration is required"})
		return
	}

	if want.Spec.Provider.Vault == nil {
		c.JSON(400, gin.H{"error": "Only Vault provider is supported"})
		return
	}

	log.G().WithFields(log.Fields{
		"name":          want.Name,
		"namespace":     want.Namespace,
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

	if err := createExternalSecret(context.Background(), &want, &AppTemplateCreateOptions{
		CloneOpts: cloneOpts,
	}); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create external secret: %v", err)})
		return
	}

	c.JSON(201, gin.H{
		"name":    want.Name,
		"message": "SecretStore created successfully",
	})
}

// create the external secret to gitops repo
func createExternalSecret(ctx context.Context, ss *esv1beta1.SecretStore, opts *AppTemplateCreateOptions) error {
	var (
		err error
	)

	log.G().WithField("cloneOpts", opts.CloneOpts).Debug("clone options")

	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return err
	}

	ssYaml, err := yaml.Marshal(ss)
	if err != nil {
		return err
	}

	ssExists := repofs.ExistsOrDie(repofs.Join(store.Default.ArgoCDName, ss.GetName()+".yaml"))
	if ssExists {
		return fmt.Errorf("secret store '%s' already exists", ss.GetName())
	}
	log.G().Debug("repository is ok")

	bulkWrites := []fs.BulkWriteRequest{}
	bulkWrites = append(bulkWrites, fs.BulkWriteRequest{
		Filename: repofs.Join(store.Default.ArgoCDName, ss.GetName()+".yaml"),
		Data:     util.JoinManifests(ssYaml),
		ErrMsg:   "failed to create secret store file",
	})

	if err = fs.BulkWrite(repofs, bulkWrites...); err != nil {
		return err
	}

	log.G().Infof("pushing new secret store manifest to repo")
	if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: fmt.Sprintf("Added secret store '%s'", ss.GetName())}); err != nil {
		return err
	}

	log.G().Infof("secret store created: '%s'", ss.GetName())

	return nil
}
