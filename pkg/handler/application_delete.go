package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/application/repotarget"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"
)

func DeleteApplicationHandler(c *gin.Context) {
	username := c.GetString(middleware.UserNameKey)
	tenant := c.GetString(middleware.TenantKey)
	appName := c.Param("name")

	log.G().WithFields(log.Fields{
		"username": username,
		"tenant":   tenant,
		"appName":  appName,
	}).Debug("delete argo application")

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	applicationName := fmt.Sprintf("%s-%s", tenant, appName)
	_, err = argoClient.Applications(store.Default.ArgoCDNamespace).Get(context.Background(), applicationName, metav1.GetOptions{})
	if err != nil {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Application not found: %v", err)})
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

	var native repotarget.NativeRepoTarget

	if err := native.RunAppDelete(context.Background(), &types.AppDeleteOptions{
		CloneOpts:   cloneOpts,
		ProjectName: tenant,
		AppName:     appName,
		Global:      false,
	}); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete application: %v", err)})
		return
	}

	c.JSON(204, nil)
}
