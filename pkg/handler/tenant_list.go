package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/store"
)

type ProjectInfo struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	DefaultCluster string `json:"default_cluster"`
}

func ListProjects(c *gin.Context) {
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

	projects, err := RunProjectList(context.Background(), &ProjectListOptions{CloneOpts: cloneOpts})
	if err != nil {
		log.Errorf("Failed to list projects: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list projects: %v", err)})
		return
	}

	c.JSON(200, gin.H{"projects": projects})
}

func RunProjectList(ctx context.Context, opts *ProjectListOptions) ([]ProjectInfo, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return nil, err
	}

	matches, err := billyUtils.Glob(repofs, repofs.Join(store.Default.ProjectsDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	var projects []ProjectInfo

	for _, name := range matches {
		proj, _, err := getProjectInfoFromFile(repofs, name)
		if err != nil {
			return nil, err
		}

		projectInfo := ProjectInfo{
			Name:           proj.Name,
			Namespace:      proj.Namespace,
			DefaultCluster: proj.Annotations[store.Default.DestServerAnnotation],
		}
		projects = append(projects, projectInfo)
	}

	return projects, nil
}

var getProjectInfoFromFile = func(repofs fs.FS, name string) (*argocdv1alpha1.AppProject, *argocdv1alpha1.ApplicationSet, error) {
	proj := &argocdv1alpha1.AppProject{}
	appSet := &argocdv1alpha1.ApplicationSet{}
	if err := repofs.ReadYamls(name, proj, appSet); err != nil {
		return nil, nil, err
	}

	return proj, appSet, nil
}
