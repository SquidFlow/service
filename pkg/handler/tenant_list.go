package handler

import (
	"context"
	"fmt"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
)

type TenantInfo struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	DefaultCluster string `json:"default_cluster"`
}

func ListTenants(c *gin.Context) {
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

	tenants, err := RunProjectList(context.Background(), &ProjectListOptions{CloneOpts: cloneOpts})
	if err != nil {
		log.G().Errorf("Failed to list tenants: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list tenants: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"total":   len(tenants),
		"items":   tenants,
	})
}

func RunProjectList(ctx context.Context, opts *ProjectListOptions) ([]TenantInfo, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return nil, err
	}

	matches, err := billyUtils.Glob(repofs, repofs.Join(store.Default.ProjectsDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	var tenants []TenantInfo

	for _, name := range matches {
		proj, _, err := getProjectInfoFromFile(repofs, name)
		if err != nil {
			return nil, err
		}

		tenantInfo := TenantInfo{
			Name:           proj.Name,
			Namespace:      proj.Namespace,
			DefaultCluster: proj.Annotations[store.Default.DestServerAnnotation],
		}
		tenants = append(tenants, tenantInfo)
	}

	return tenants, nil
}

var getProjectInfoFromFile = func(repofs fs.FS, name string) (*argocdv1alpha1.AppProject, *argocdv1alpha1.ApplicationSet, error) {
	proj := &argocdv1alpha1.AppProject{}
	appSet := &argocdv1alpha1.ApplicationSet{}
	if err := repofs.ReadYamls(name, proj, appSet); err != nil {
		return nil, nil, err
	}

	return proj, appSet, nil
}
