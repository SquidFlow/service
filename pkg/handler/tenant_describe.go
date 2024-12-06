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
	"github.com/squidflow/service/pkg/middleware"
)

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

	var nativeRepoWriter = repotarget.NativeRepoTarget{}
	tenantResp, err := nativeRepoWriter.RunProjectGetDetail(context.Background(), projectName, cloneOpts)
	if err != nil {
		log.G().Errorf("Failed to get project detail: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get project detail: %v", err)})
		return
	}

	c.JSON(200, tenantResp)
}
