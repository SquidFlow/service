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

	opts := &types.ProjectDeleteOptions{
		CloneOpts:   cloneOpts,
		ProjectName: projectName,
	}

	var nativeRepoWriter = repotarget.NativeRepoTarget{}
	err := nativeRepoWriter.RunProjectDelete(context.Background(), opts)
	if err != nil {
		log.G().Errorf("Failed to delete project: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete project: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Project '%s' deleted successfully", projectName)})
}
