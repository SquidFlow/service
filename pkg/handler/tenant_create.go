package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/application/repotarget"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

func CreateTenant(c *gin.Context) {
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

	var nativeRepoWriter = repotarget.NativeRepoTarget{}
	err := nativeRepoWriter.RunProjectCreate(context.Background(), opts)
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
