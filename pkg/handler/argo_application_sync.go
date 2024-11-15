package handler

import (
	"context"
	"fmt"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/h4-poc/service/pkg/kube"
)

type SyncOptions struct {
	Prune        bool `json:"prune"`
	ValidateOnly bool `json:"validateOnly"`
}

// SyncRequest represents the request structure for syncing applications
type SyncRequest struct {
	Applications []string    `json:"applications" binding:"required,min=1"`
	Options      SyncOptions `json:"options"`
}

// SyncResponse represents the response structure for sync operation
type SyncResponse struct {
	Results []ApplicationSyncResult `json:"results"`
}

// ApplicationSyncResult represents the sync result for a single application
type ApplicationSyncResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// TODO: fix this function
// SyncArgoApplication handles the synchronization of one or more Argo CD applications
func SyncArgoApplication(c *gin.Context) {
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
		return
	}

	// Create ArgoCD client
	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	response := SyncResponse{
		Results: make([]ApplicationSyncResult, 0, len(req.Applications)),
	}

	// Process each application
	for _, appName := range req.Applications {
		result := ApplicationSyncResult{
			Name: appName,
		}

		// Get the application
		app, err := argoClient.Applications(appName).Get(context.Background(), appName, metav1.GetOptions{})
		if err != nil {
			result.Status = "Failed"
			result.Message = fmt.Sprintf("Failed to get application: %v", err)
			response.Results = append(response.Results, result)
			continue
		}

		// Prepare sync operation
		syncOp := argocdv1alpha1.SyncOperation{
			Prune:  req.Options.Prune,
			DryRun: req.Options.ValidateOnly,
		}

		// Update application with sync operation
		app.Operation = &argocdv1alpha1.Operation{
			Sync: &syncOp,
		}

		// Apply the sync operation
		if err != nil {
			result.Status = "Failed"
			result.Message = fmt.Sprintf("Failed to sync application: %v", err)
		} else {
			result.Status = "Syncing"
			result.Message = "Application sync initiated successfully"
		}

		response.Results = append(response.Results, result)
		log.WithFields(log.Fields{
			"application": appName,
			"status":      result.Status,
			"message":     result.Message,
		}).Info("Application sync result")
	}

	c.JSON(200, response)
}

// validateSyncRequest validates the sync request
func validateSyncRequest(req *SyncRequest) error {
	if len(req.Applications) == 0 {
		return fmt.Errorf("at least one application must be specified")
	}
	return nil
}
