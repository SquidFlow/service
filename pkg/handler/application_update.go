package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
)

type UpdateOptions struct {
	CloneOpts   *git.CloneOptions
	ProjectName string
	AppName     string
	Username    string
	UpdateReq   *ApplicationUpdateRequest
	KubeFactory kube.Factory
	Annotations map[string]string
}

func UpdateApplicationHandler(c *gin.Context) {
	username := c.GetString(middleware.UserNameKey)
	tenant := c.GetString(middleware.TenantKey)
	appName := c.Param("name")

	log.G().WithFields(log.Fields{
		"username": username,
		"tenant":   tenant,
		"appName":  appName,
	}).Debug("update argo application")

	var updateReq ApplicationUpdateRequest
	if err := c.BindJSON(&updateReq); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
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

	annotations := make(map[string]string)
	if updateReq.ApplicationInstantiation.Description != "" {
		annotations["squidflow.github.io/description"] = updateReq.ApplicationInstantiation.Description
	}

	// TODO: support security
	log.G().WithFields(log.Fields{
		"security": updateReq.ApplicationInstantiation.Security,
	}).Debug("TODO support security")

	// TODO: support ingress
	log.G().WithFields(log.Fields{
		"ingress": updateReq.ApplicationInstantiation.Ingress,
	}).Debug("TODO support ingress")

	annotations["squidflow.github.io/last-modified-by"] = username

	updateOpts := &UpdateOptions{
		CloneOpts:   cloneOpts,
		ProjectName: tenant,
		AppName:     appName,
		Username:    username,
		UpdateReq:   &updateReq,
		KubeFactory: kube.NewFactory(),
		Annotations: annotations,
	}

	if err := updateApplication(context.Background(), updateOpts); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update application: %v", err)})
		return
	}

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	app, err := getApplicationDetail(context.Background(), &AppListOptions{
		CloneOpts:    cloneOpts,
		ProjectName:  tenant,
		ArgoCDClient: argoClient,
	}, appName)

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get updated application details: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message":     "Application updated successfully",
		"application": app,
	})
}

// TODO: implement this function
func updateApplication(ctx context.Context, opts *UpdateOptions) error {
	return nil
}
