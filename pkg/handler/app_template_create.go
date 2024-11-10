package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"

	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/log"
)

// CreateApplicationTemplateRequest defines the request body for creating an ApplicationTemplate
type CreateApplicationTemplateRequest struct {
	Name        string            `json:"name" binding:"required"`
	Path        string            `json:"path" binding:"required"`
	Owner       string            `json:"owner" binding:"required"`
	Source      ApplicationSource `json:"source" binding:"required"`
	Description string            `json:"description,omitempty"`
	AppType     string            `json:"appType" binding:"required,oneof=kustomize helm"`
}

type CreateApplicationTemplateResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

// CreateApplicationTemplate handles HTTP requests to create an ApplicationTemplate
func CreateApplicationTemplate(c *gin.Context) {
	// Get kube factory from context
	factory := c.MustGet("kubeFactory").(kube.Factory)

	// Get kubernetes client
	restConfig, err := factory.ToRESTConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get kubernetes client: %v", err)})
		return
	}

	k8sClient, err := client.New(restConfig, client.Options{})

	var req CreateApplicationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Create the application template
	err = createApplicationTemplate(context.Background(), k8sClient, &req)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			c.JSON(409, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create application template: %v", err)})
		return
	}

	// Return success response
	c.JSON(201, CreateApplicationTemplateResponse{
		ID:      1, // You may want to generate or retrieve a real ID
		Name:    req.Name,
		Success: true,
	})
}

// CreateApplicationTemplate creates a new ApplicationTemplate
func createApplicationTemplate(ctx context.Context, k8sClient client.Client, userReq *CreateApplicationTemplateRequest) error {
	log.G().WithField("user input request", userReq).Info("CreateApplicationTemplateRequest")

	// Create new ApplicationTemplate object
	template := &apptempv1alpha1.ApplicationTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      userReq.Name,
			Namespace: "argocd",
		},
	}

	// Check if template already exists
	existing := &apptempv1alpha1.ApplicationTemplate{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "argocd", Name: userReq.Name}, existing)
	if err == nil {
		return fmt.Errorf("application template %s already exists in namespace %s", userReq.Name, "argocd")
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check existing template: %w", err)
	}

	// for not found, create the template
	template = &apptempv1alpha1.ApplicationTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      userReq.Name,
			Namespace: "argocd",
		},
		Spec: apptempv1alpha1.ApplicationTemplateSpec{
			Name:           userReq.Name,
			RepoURL:        userReq.Source.URL,
			TargetRevision: userReq.Source.Branch,
			Helm: &apptempv1alpha1.HelmConfig{
				Chart: userReq.AppType,
			},
			Kustomize: &apptempv1alpha1.KustomizeConfig{
				RenderTargets: []apptempv1alpha1.KustomizeRenderTarget{
					{
						Path: userReq.Path,
						DestinationCluster: apptempv1alpha1.ClusterSelector{
							Name: userReq.Owner,
							MatchLabels: map[string]string{
								"env": userReq.Owner,
							},
						},
					},
				},
			},
		},
	}
	if err := k8sClient.Create(ctx, template); err != nil {
		return fmt.Errorf("failed to create application template: %w", err)
	}

	return nil
}
