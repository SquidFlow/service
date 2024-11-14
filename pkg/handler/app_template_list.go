package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"

	"github.com/h4-poc/service/pkg/kube"
)

// ListTemplateResponse represents the response structure for template listing
type ListTemplateResponse struct {
	Success bool                  `json:"success"`
	Total   int                   `json:"total"`
	Items   []ApplicationTemplate `json:"items"`
}

// ListTemplateFilter represents the filter criteria for listing templates
type ListTemplateFilter struct {
	AppType   string
	Owner     string
	Validated string
}

// ListApplicationTemplate handles the retrieval of application templates
func ListApplicationTemplate(c *gin.Context) {
	// Get query parameters for filtering
	appType := c.Query("appType")     // Filter by application type (helm/kustomize)
	owner := c.Query("owner")         // Filter by owner
	validated := c.Query("validated") // Filter by validation status

	log.WithFields(log.Fields{
		"appType":   appType,
		"owner":     owner,
		"validated": validated,
	}).Info("Listing application templates")

	// Get kube factory from context
	factory := c.MustGet("kubeFactory").(kube.Factory)

	// Get kubernetes client
	restConfig, err := factory.ToRESTConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get kubernetes client: %v", err)})
		return
	}

	k8sClient, err := client.New(restConfig, client.Options{})

	appTempList := apptempv1alpha1.ApplicationTemplateList{}

	err = k8sClient.List(context.Background(), &appTempList, &client.ListOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list application templates: %v", err)})
		return
	}

	if err != nil {
		c.JSON(500, ListTemplateResponse{
			Success: false,
			Total:   0,
			Items:   nil,
		})
		return
	}

	// Return success response
	c.JSON(200, ListTemplateResponse{
		Success: true,
		Total:   len(appTempList.Items),
		Items:   convertApplicationTemplate(&appTempList),
	})
}

// convertApplicationTemplate converts an ApplicationTemplate to a handler ApplicationTemplate
func convertApplicationTemplate(template *apptempv1alpha1.ApplicationTemplateList) []ApplicationTemplate {
	var ret []ApplicationTemplate
	for _, item := range template.Items {
		ret = append(ret, ApplicationTemplate{
			Name:        item.Name,
			Owner:       item.Annotations["owner"],
			Description: item.Annotations["description"],
			AppType:     getAppTempType(item),
			Source: ApplicationSource{
				URL:            item.Spec.RepoURL,
				TargetRevision: item.Spec.TargetRevision,
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
		})
	}
	return ret
}

func getAppTempType(temp apptempv1alpha1.ApplicationTemplate) string {
	var enableHelm, enableKustomize bool
	if temp.Spec.Helm != nil {
		enableHelm = true
	}
	if temp.Spec.Kustomize != nil {
		enableKustomize = true
	}

	// only define 2 types: helm and kustomize
	if enableHelm && enableKustomize {
		return "kustomize"
	}
	if enableHelm {
		return "helm"
	}
	if enableKustomize {
		return "kustomize"
	}
	return "unknown"
}
