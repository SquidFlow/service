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

type DeleteTemplateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func DeleteApplicationTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		c.JSON(400, gin.H{"error": "template ID is required"})
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

	if err := RunDeleteApplicationTemplate(context.Background(), templateID, &AppTemplateDeleteOptions{
		CloneOpts: cloneOpts,
	}); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete template: %v", err)})
		return
	}

	c.JSON(200, DeleteTemplateResponse{
		Success: true,
		Message: "template deleted successfully",
	})
}

type AppTemplateDeleteOptions struct {
	CloneOpts *git.CloneOptions
}

func RunDeleteApplicationTemplate(ctx context.Context, templateID string, opts *AppTemplateDeleteOptions) error {
	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return err
	}

	templatePath := repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		fmt.Sprintf("apptemp-%s.yaml", templateID),
	)

	exists := repofs.ExistsOrDie(templatePath)
	if !exists {
		log.G().Infof("template %s not found, considering it as already deleted", templateID)
		return nil
	}

	if err := repofs.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to delete template file: %v", err)
	}

	if _, err = r.Persist(ctx, &git.PushOptions{
		CommitMsg: fmt.Sprintf("chore: deleted app template '%s'", templateID),
	}); err != nil {
		return fmt.Errorf("failed to push template deletion to repo: %v", err)
	}

	log.G().Infof("app template deleted: '%s'", templateID)
	return nil
}
