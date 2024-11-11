package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"
	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

// UpdateTemplateRequest represents the request for updating an application template
type UpdateTemplateRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Path        string                  `json:"path" binding:"required"`
	Owner       string                  `json:"owner" binding:"required"`
	Source      ApplicationSource       `json:"source" binding:"required"`
	Description string                  `json:"description,omitempty"`
	AppType     ApplicationTemplateType `json:"appType" binding:"required,oneof=kustomize helm"`
}

// UpdateTemplateResponse represents the response for template update operation
type UpdateTemplateResponse struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
}

func UpdateApplicationTemplate(c *gin.Context) {
	// Get template ID from path parameter
	templateName := c.Param("template_id")
	if templateName == "" {
		c.JSON(400, gin.H{"error": "template name is required"})
		return
	}

	// Get kube factory from context
	factory := c.MustGet("kubeFactory").(kube.Factory)

	// Get kubernetes client
	restConfig, err := factory.ToRESTConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get kubernetes client: %v", err)})
		return
	}

	k8sClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create kubernetes client: %v", err)})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Update the template
	err = updateApplicationTemplate(context.Background(), k8sClient, templateName, &req)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Template not found"})
			return
		}
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update template: %v", err)})
		return
	}

	c.JSON(200, UpdateTemplateResponse{
		Success: true,
		Name:    req.Name,
	})
}

func updateApplicationTemplate(ctx context.Context, k8sClient client.Client, templateName string, req *UpdateTemplateRequest) error {
	log.G().WithField("template", templateName).Info("Updating application template")

	// Get existing template
	existing := &apptempv1alpha1.ApplicationTemplate{}
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: store.Default.ArgoCDNamespace,
		Name:      templateName,
	}, existing)
	if err != nil {
		return err
	}

	// Update template fields
	existing.Spec.Name = req.Name
	existing.Spec.RepoURL = req.Source.URL
	existing.Spec.TargetRevision = req.Source.TargetRevision

	if req.AppType == "helm" {
		existing.Spec.Helm = &apptempv1alpha1.HelmConfig{
			Chart:      req.Name,
			Version:    "v1",
			Repository: req.Source.URL,
			RenderTargets: []apptempv1alpha1.HelmRenderTarget{
				{
					DestinationCluster: apptempv1alpha1.ClusterSelector{
						Name: req.Owner,
						MatchLabels: map[string]string{
							"env": req.Owner,
						},
					},
					ValuesPath: req.Path,
				},
			},
		}
		existing.Spec.Kustomize = nil
	} else {
		existing.Spec.Kustomize = &apptempv1alpha1.KustomizeConfig{
			RenderTargets: []apptempv1alpha1.KustomizeRenderTarget{
				{
					Path: req.Path,
					DestinationCluster: apptempv1alpha1.ClusterSelector{
						Name: req.Owner,
						MatchLabels: map[string]string{
							"env": req.Owner,
						},
					},
				},
			},
		}
		existing.Spec.Helm = nil
	}

	// Update the template
	if err := k8sClient.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update application template: %w", err)
	}

	return nil
}
