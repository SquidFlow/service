package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

// DeleteProject handles the HTTP request to delete a project
func DeleteTenant(c *gin.Context) {
	projectName := c.Param("name")
	if projectName == "" {
		c.JSON(400, gin.H{"error": "Project name is required"})
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

	opts := &ProjectDeleteOptions{
		CloneOpts:   cloneOpts,
		ProjectName: projectName,
	}

	err := RunProjectDelete(context.Background(), opts)
	if err != nil {
		log.G().Errorf("Failed to delete project: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete project: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Project '%s' deleted successfully", projectName)})
}

func RunProjectDelete(ctx context.Context, opts *ProjectDeleteOptions) error {
	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return err
	}

	allApps, err := repofs.ReadDir(store.Default.AppsDir)
	if err != nil {
		return fmt.Errorf("failed to list all applications")
	}

	for _, app := range allApps {
		err = application.DeleteFromProject(repofs, app.Name(), opts.ProjectName)
		if err != nil {
			return err
		}
	}

	err = repofs.Remove(repofs.Join(store.Default.ProjectsDir, opts.ProjectName+".yaml"))
	if err != nil {
		return fmt.Errorf("failed to delete project '%s': %w", opts.ProjectName, err)
	}

	log.G().Info("committing changes to gitops repo...")
	if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: fmt.Sprintf("chore: deleted project '%s'", opts.ProjectName)}); err != nil {
		return fmt.Errorf("failed to push to repo: %w", err)
	}

	return nil
}
