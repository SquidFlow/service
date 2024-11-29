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

type ApplicationUpdate struct {
	Description         string              `json:"description,omitempty"`
	DestinationClusters DestinationClusters `json:"destination_clusters,omitempty"`
	Ingress             *Ingress            `json:"ingress,omitempty"`
	Security            *Security           `json:"security,omitempty"`
}

type UpdateOptions struct {
	CloneOpts   *git.CloneOptions
	ProjectName string
	AppName     string
	Username    string
	UpdateReq   *ApplicationUpdate
	KubeFactory kube.Factory
	Annotations map[string]string
}

func UpdateArgoApplication(c *gin.Context) {
	username := c.GetString(middleware.UserNameKey)
	tenant := c.GetString(middleware.TenantKey)
	appName := c.Param("name")

	log.G().WithFields(log.Fields{
		"username": username,
		"tenant":   tenant,
		"appName":  appName,
	}).Debug("update argo application")

	var updateReq ApplicationUpdate
	if err := c.BindJSON(&updateReq); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if err := validateUpdateRequest(&updateReq); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
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
	if updateReq.Description != "" {
		annotations["squidflow.github.io/description"] = updateReq.Description
	}
	if updateReq.Ingress != nil {
		annotations["squidflow.github.io/ingress.host"] = updateReq.Ingress.Host
		if updateReq.Ingress.TLS != nil {
			annotations["squidflow.github.io/ingress.tls.enabled"] = fmt.Sprintf("%v", updateReq.Ingress.TLS.Enabled)
			annotations["squidflow.github.io/ingress.tls.secretName"] = updateReq.Ingress.TLS.SecretName
		}
	}
	if updateReq.Security != nil && updateReq.Security.ExternalSecret != nil {
		annotations["squidflow.github.io/security.external_secret.secret_store_ref.id"] = updateReq.Security.ExternalSecret.SecretStoreRef.ID
	}
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

func updateApplication(ctx context.Context, opts *UpdateOptions) error {
	return nil
}

func validateUpdateRequest(update *ApplicationUpdate) error {
	if update.Ingress != nil {
		if err := validateIngress(update.Ingress); err != nil {
			return fmt.Errorf("invalid ingress configuration: %w", err)
		}
	}

	if update.Security != nil {
		if err := validateSecurity(update.Security); err != nil {
			return fmt.Errorf("invalid security configuration: %w", err)
		}
	}

	if update.DestinationClusters.Clusters != nil {
		if err := validateDestinationClusters(&update.DestinationClusters); err != nil {
			return fmt.Errorf("invalid destination_clusters: %w", err)
		}
	}

	return nil
}
