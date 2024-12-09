package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/application/repowriter"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/types"
)

func TenantCreate(c *gin.Context) {
	var req types.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
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

	opts := &types.ProjectCreateOptions{
		CloneOpts:   cloneOpts,
		ProjectName: req.ProjectName,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	}

	err := repowriter.GetRepoWriter().RunProjectCreate(context.Background(), opts)
	if err != nil {
		log.G().Errorf("Failed to create project: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create project: %v", err)})
		return
	}

	c.JSON(201, gin.H{
		"message": fmt.Sprintf("Project '%s' created successfully", req.ProjectName),
		"project": req,
	})
}

func TenantDelete(c *gin.Context) {
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

	opts := &types.ProjectDeleteOptions{
		CloneOpts:   cloneOpts,
		ProjectName: projectName,
	}

	err := repowriter.GetRepoWriter().RunProjectDelete(context.Background(), opts)
	if err != nil {
		log.G().Errorf("Failed to delete project: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete project: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Project '%s' deleted successfully", projectName)})
}

func TenantGet(c *gin.Context) {
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

	tenantResp, err := repowriter.GetRepoWriter().RunProjectGetDetail(context.Background(), projectName, cloneOpts)
	if err != nil {
		log.G().Errorf("Failed to get project detail: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get project detail: %v", err)})
		return
	}

	c.JSON(200, tenantResp)
}

func TenantsList(c *gin.Context) {
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

	tenants, err := repowriter.GetRepoWriter().RunProjectList(context.Background(), &types.ProjectListOptions{CloneOpts: cloneOpts})
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
