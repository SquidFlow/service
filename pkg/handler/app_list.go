package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/store"
)

type AppListOptions struct {
	CloneOpts   *git.CloneOptions
	ProjectName string
}

type AppListResponse struct {
	ProjectName   string `json:"project_name"`
	Name          string `json:"name"`
	DestNamespace string `json:"dest_namespace"`
	DestServer    string `json:"dest_server"`
}

func ListApplications(c *gin.Context) {
	var project string
	if c.Query("project") != "" {
		project = c.Query("project")
	}

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

	apps, err := RunAppList(context.Background(), &AppListOptions{CloneOpts: cloneOpts, ProjectName: project})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list applications: %v", err)})
		return
	}

	c.JSON(200, gin.H{"applications": apps})
}

func RunAppList(ctx context.Context, opts *AppListOptions) ([]AppListResponse, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return nil, err
	}

	// get all apps beneath apps/*/overlays/<project>
	matches, err := billyUtils.Glob(repofs, repofs.Join(store.Default.AppsDir, "*", store.Default.OverlaysDir, opts.ProjectName))
	if err != nil {
		return nil, fmt.Errorf("failed to run glob on %s: %w", opts.ProjectName, err)
	}

	var apps []AppListResponse

	for _, appPath := range matches {
		conf, err := getConfigFileFromPath(repofs, appPath)
		if err != nil {
			return nil, err
		}

		app := AppListResponse{
			ProjectName:   opts.ProjectName,
			Name:          conf.UserGivenName,
			DestNamespace: conf.DestNamespace,
			DestServer:    conf.DestServer,
		}
		apps = append(apps, app)
	}

	return apps, nil
}
