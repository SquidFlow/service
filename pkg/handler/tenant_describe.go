package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
)

type TenantDetailInfo struct {
	Name                       string            `json:"name"`
	Namespace                  string            `json:"namespace"`
	Description                string            `json:"description,omitempty"`
	DefaultCluster             string            `json:"default_cluster"`
	SourceRepos                []string          `json:"source_repos,omitempty"`
	Destinations               []ProjectDest     `json:"destinations,omitempty"`
	ClusterResourceWhitelist   []ProjectResource `json:"cluster_resource_whitelist,omitempty"`
	NamespaceResourceWhitelist []ProjectResource `json:"namespace_resource_whitelist,omitempty"`
	CreatedBy                  string            `json:"created_by"`
	CreatedAt                  string            `json:"created_at,omitempty"`
}

type ProjectDest struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

type ProjectResource struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
}

func DescribeTenant(c *gin.Context) {
	tenant := c.GetString(middleware.TenantKey)
	log.G().Infof("auth context info tenant: %s", tenant)

	projectName := c.Param("name")

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

	tenantResp, err := getProjectDetail(context.Background(), projectName, cloneOpts)
	if err != nil {
		log.G().Errorf("Failed to get project detail: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get project detail: %v", err)})
		return
	}

	c.JSON(200, tenantResp)
}

func getProjectDetail(ctx context.Context, projectName string, opts *git.CloneOptions) (*TenantDetailInfo, error) {
	_, repofs, err := prepareRepo(ctx, opts, projectName)
	if err != nil {
		return nil, err
	}

	projectFile := repofs.Join(store.Default.ProjectsDir, projectName+".yaml")
	if !repofs.ExistsOrDie(projectFile) {
		return nil, fmt.Errorf("project %s not found", projectName)
	}

	proj, _, err := getProjectInfoFromFile(repofs, projectFile)
	if err != nil {
		return nil, err
	}

	detail := &TenantDetailInfo{
		Name:           proj.Name,
		Namespace:      proj.Namespace,
		Description:    proj.Annotations["description"],
		DefaultCluster: proj.Annotations[store.Default.DestServerAnnotation],
		CreatedBy:      proj.Annotations["created-by"],
		CreatedAt:      proj.CreationTimestamp.String(),
	}

	if len(proj.Spec.SourceRepos) > 0 {
		detail.SourceRepos = proj.Spec.SourceRepos
	}

	if len(proj.Spec.Destinations) > 0 {
		for _, dest := range proj.Spec.Destinations {
			detail.Destinations = append(detail.Destinations, ProjectDest{
				Server:    dest.Server,
				Namespace: dest.Namespace,
			})
		}
	}

	if len(proj.Spec.ClusterResourceWhitelist) > 0 {
		for _, res := range proj.Spec.ClusterResourceWhitelist {
			detail.ClusterResourceWhitelist = append(detail.ClusterResourceWhitelist, ProjectResource{
				Group: res.Group,
				Kind:  res.Kind,
			})
		}
	}

	if len(proj.Spec.NamespaceResourceWhitelist) > 0 {
		for _, res := range proj.Spec.NamespaceResourceWhitelist {
			detail.NamespaceResourceWhitelist = append(detail.NamespaceResourceWhitelist, ProjectResource{
				Group: res.Group,
				Kind:  res.Kind,
			})
		}
	}

	return detail, nil
}
