package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/store"
)

// UpdateTemplateRequest represents the request for updating an application template
type UpdateTemplateRequest struct {
	Name        string                  `json:"name,omitempty"`
	Path        string                  `json:"path,omitempty"`
	Owner       string                  `json:"owner,omitempty"`
	Source      *ApplicationSource      `json:"source,omitempty"`
	Description string                  `json:"description,omitempty"`
	AppType     ApplicationTemplateType `json:"appType,omitempty"`
}

// UpdateTemplateResponse represents the response for template update operation
type UpdateTemplateResponse struct {
	Item    ApplicationTemplate `json:"item"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
}

func UpdateApplicationTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		c.JSON(400, gin.H{"error": "template ID is required"})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Get kube factory from context
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

	_, repofs, err := prepareRepo(context.Background(), cloneOpts, "")
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get application template: %v", err)})
		return
	}

	templatePath := repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		fmt.Sprintf("apptemp-%s.yaml", templateID),
	)

	template := &apptempv1alpha1.ApplicationTemplate{}
	if err := repofs.ReadYamls(templatePath, template); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get application template: %v", err)})
		return
	}

	if req.Name != "" {
		template.Name = req.Name
	}

	if req.Source != nil {
		if req.Source.URL != "" {
			template.Spec.RepoURL = req.Source.URL
		}
		if req.Source.TargetRevision != "" {
			template.Spec.TargetRevision = req.Source.TargetRevision
		}
	}
	if req.Owner != "" {
		template.Annotations["h4-poc.github.io/owner"] = req.Owner
	}
	if req.Description != "" {
		template.Annotations["h4-poc.github.io/description"] = req.Description
	}
	template.Annotations["h4-poc.github.io/updated-at"] = time.Now().Format(time.RFC3339)

	if req.AppType != "" {
		switch req.AppType {
		case ApplicationTemplateTypeHelm:
			template.Spec.Helm = &apptempv1alpha1.HelmConfig{
				Chart:      template.Name,
				Version:    "v1",
				Repository: template.Spec.RepoURL,
				RenderTargets: []apptempv1alpha1.HelmRenderTarget{
					{
						DestinationCluster: apptempv1alpha1.ClusterSelector{
							Name: template.Annotations["h4-poc.github.io/owner"],
							MatchLabels: map[string]string{
								"env": template.Annotations["h4-poc.github.io/owner"],
							},
						},
						ValuesPath: req.Path,
					},
				},
			}
			template.Spec.Kustomize = nil
		case ApplicationTemplateTypeKustomize:
			template.Spec.Kustomize = &apptempv1alpha1.KustomizeConfig{
				RenderTargets: []apptempv1alpha1.KustomizeRenderTarget{
					{
						Path: req.Path,
						DestinationCluster: apptempv1alpha1.ClusterSelector{
							Name: template.Annotations["h4-poc.github.io/owner"],
							MatchLabels: map[string]string{
								"env": template.Annotations["h4-poc.github.io/owner"],
							},
						},
					},
				},
			}
			template.Spec.Helm = nil
		}
	}

	err = writeAppTemplate2Repo(
		context.Background(),
		template,
		&AppTemplateCreateOptions{
			CloneOpts: cloneOpts,
		},
	)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update template: %v", err)})
		return
	}

	c.JSON(200, UpdateTemplateResponse{
		Item: ApplicationTemplate{
			Name:        template.Name,
			Owner:       template.Annotations["h4-poc.github.io/owner"],
			Description: template.Annotations["h4-poc.github.io/description"],
			ID:          template.Annotations["h4-poc.github.io/id"],
			CreatedAt:   template.Annotations["h4-poc.github.io/created-at"],
			UpdatedAt:   template.Annotations["h4-poc.github.io/updated-at"],
			AppType:     getAppTempType(*template),
			Validated:   true,
			Path:        template.Spec.Helm.RenderTargets[0].ValuesPath,
			Source: ApplicationSource{
				URL:            template.Spec.RepoURL,
				TargetRevision: template.Spec.TargetRevision,
			},
			Resources: ApplicationResources{
				Deployments: 2,
				Services:    1,
				Configmaps:  1,
			},
			Events: []ApplicationEvent{
				{
					Time: "2021-09-01T00:00:00Z",
					Type: "Normal",
				},
			},
		},
		Success: true,
		Message: "template updated successfully",
	})
}
