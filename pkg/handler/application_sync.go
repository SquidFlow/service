package handler

import (
	"context"
	"fmt"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/kube"
)

// SyncApplicationHandler handles the synchronization of one or more Argo CD applications
func SyncApplicationHandler(c *gin.Context) {
	var req SyncApplicationRequest
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

	response := SyncApplicationResponse{
		Results: make([]SyncApplicationResult, 0, len(req.Applications)),
	}

	// Process each application
	for _, appName := range req.Applications {
		result := SyncApplicationResult{
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
			Prune: false,
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
