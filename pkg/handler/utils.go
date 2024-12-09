package handler

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/google/uuid"
)

// getNewId returns a new id for the resource
func getNewId() string {
	return uuid.New().String()
}

// getAppStatus returns the status of the ArgoCD application
func getAppStatus(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}

	// Check if OperationState exists and has Phase
	if app.Status.OperationState != nil && app.Status.OperationState.Phase != "" {
		return string(app.Status.OperationState.Phase)
	}

	// If no OperationState, try to get status from Sync
	if app.Status.Sync.Status != "" {
		return string(app.Status.Sync.Status)
	}

	// Default status if nothing else is available
	return "Unknown"
}

// getAppHealth returns the health status of the ArgoCD application
func getAppHealth(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}

	// HealthStatus is a struct, we should check if it's empty instead
	if app.Status.Health.Status == "" {
		return "Unknown"
	}
	return string(app.Status.Health.Status)
}

// getAppSyncStatus returns the sync status of the ArgoCD application
func getAppSyncStatus(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}
	return string(app.Status.Sync.Status)
}
