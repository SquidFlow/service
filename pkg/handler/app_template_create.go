package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

// CreateApplicationTemplateRequest defines the request body for creating an ApplicationTemplate
type CreateApplicationTemplateRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Path        string                  `json:"path" binding:"required"`
	Owner       string                  `json:"owner" binding:"required"`
	Source      ApplicationSource       `json:"source" binding:"required"`
	Description string                  `json:"description,omitempty"`
	AppType     ApplicationTemplateType `json:"appType" binding:"required,oneof=kustomize helm"`
}

type CreateApplicationTemplateResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

// CreateApplicationTemplate handles HTTP requests to create an ApplicationTemplate
func CreateApplicationTemplate(c *gin.Context) {
	dynamicClient := c.MustGet("dynamicClient").(dynamic.Interface)
	discoveryClient := c.MustGet("discoveryClient").(*discovery.DiscoveryClient)

	var req CreateApplicationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	err := createApplicationTemplate(context.Background(), dynamicClient, discoveryClient, &req)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			c.JSON(409, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create application template: %v", err)})
		return
	}

	c.JSON(201, CreateApplicationTemplateResponse{
		ID:      1,
		Name:    req.Name,
		Success: true,
	})
}

func createApplicationTemplate(ctx context.Context, dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, userReq *CreateApplicationTemplateRequest) error {
	log.G().WithField("user input request", userReq).Info("create application template...")

	_, resourceList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return fmt.Errorf("failed to get server resources: %w", err)
	}

	appTemplateGVR := schema.GroupVersionResource{
		Group:    "argocd-addon.github.com",
		Version:  "v1alpha1",
		Resource: "applicationtemplates",
	}

	resourceExists := false
	for _, list := range resourceList {
		for _, r := range list.APIResources {
			if r.Name == appTemplateGVR.Resource {
				resourceExists = true
				break
			}
		}
	}

	if !resourceExists {
		return fmt.Errorf("ApplicationTemplate CRD is not installed in the cluster")
	}

	_, err = dynamicClient.Resource(appTemplateGVR).Namespace("argocd").Get(ctx, userReq.Name, metav1.GetOptions{})
	if err == nil {
		return fmt.Errorf("application template %s already exists in namespace %s", userReq.Name, "argocd")
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check existing template: %w", err)
	}

	template := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argocd-addon.github.com/v1alpha1",
			"kind":       "ApplicationTemplate",
			"metadata": map[string]interface{}{
				"name":      userReq.Name,
				"namespace": store.Default.ArgoCDNamespace,
			},
			"spec": map[string]interface{}{
				"name":           userReq.Name,
				"repoURL":        userReq.Source.URL,
				"targetRevision": userReq.Source.TargetRevision,
				"helm": map[string]interface{}{
					"chart":      userReq.Name,
					"version":    "v1",
					"repository": userReq.Source.URL,
					"renderTargets": []map[string]interface{}{
						{
							"destinationCluster": map[string]interface{}{
								"name": userReq.Owner,
								"matchLabels": map[string]interface{}{
									"env": userReq.Owner,
								},
							},
							"valuesPath": userReq.Path,
						},
					},
				},
				"kustomize": map[string]interface{}{
					"renderTargets": []map[string]interface{}{
						{
							"path": userReq.Path,
							"destinationCluster": map[string]interface{}{
								"name": userReq.Owner,
								"matchLabels": map[string]interface{}{
									"env": userReq.Owner,
								},
							},
						},
					},
				},
			},
		},
	}

	createOpts := metav1.CreateOptions{}
	_, err = dynamicClient.Resource(appTemplateGVR).Namespace(store.Default.ArgoCDNamespace).Create(ctx, template, createOpts)
	if err != nil {
		return fmt.Errorf("failed to create application template: %w", err)
	}

	return nil
}
