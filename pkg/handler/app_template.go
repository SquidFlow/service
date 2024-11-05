package handler

import (
	"fmt"
)

// ApplicationTemplate represents a template for deploying applications
type ApplicationTemplate struct {
	ID           int                  `json:"id"`
	Name         string               `json:"name"`
	Path         string               `json:"path"`
	Validated    bool                 `json:"validated"`
	Owner        string               `json:"owner"`
	Environments []string             `json:"environments"`
	LastApplied  string               `json:"lastApplied"`
	AppType      string               `json:"appType"`
	Source       ApplicationSource    `json:"source"`
	Resources    ApplicationResources `json:"resources"`
	Events       []ApplicationEvent   `json:"events"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

type ApplicationSource struct {
	Type   string `json:"type"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
}

type ApplicationResources struct {
	Deployments               int                       `json:"deployments"`
	Services                  int                       `json:"services"`
	Configmaps                int                       `json:"configmaps"`
	Secrets                   int                       `json:"secrets"`
	Ingresses                 int                       `json:"ingresses"`
	ServiceAccounts           int                       `json:"serviceAccounts"`
	Roles                     int                       `json:"roles"`
	RoleBindings              int                       `json:"roleBindings"`
	NetworkPolicies           int                       `json:"networkPolicies"`
	PersistentVolumeClaims    int                       `json:"persistentVolumeClaims"`
	HorizontalPodAutoscalers  int                       `json:"horizontalPodAutoscalers"`
	CustomResourceDefinitions CustomResourceDefinitions `json:"customResourceDefinitions"`
}

type CustomResourceDefinitions struct {
	ExternalSecrets     int `json:"externalSecrets"`
	Certificates        int `json:"certificates"`
	IngressRoutes       int `json:"ingressRoutes"`
	PrometheusRules     int `json:"prometheusRules"`
	ServiceMeshPolicies int `json:"serviceMeshPolicies"`
	VirtualServices     int `json:"virtualServices"`
}

type ApplicationEvent struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// CreateTemplateRequest represents the request for creating a new application template
type CreateTemplateRequest struct {
	Name         string            `json:"name" binding:"required"`
	Path         string            `json:"path" binding:"required"`
	Owner        string            `json:"owner" binding:"required"`
	Environments []string          `json:"environments" binding:"required"`
	AppType      string            `json:"appType" binding:"required"`
	Source       ApplicationSource `json:"source" binding:"required"`
}

// validateTemplate validates the application template
func validateTemplate(req *CreateTemplateRequest) error {
	// Validate app type
	validAppTypes := map[string]bool{
		"kustomization": true,
		"helm":          true,
	}
	if !validAppTypes[req.AppType] {
		return fmt.Errorf("invalid app type: %s", req.AppType)
	}

	// Validate source type
	validSourceTypes := map[string]bool{
		"git": true,
	}
	if !validSourceTypes[req.Source.Type] {
		return fmt.Errorf("invalid source type: %s", req.Source.Type)
	}

	// Validate environments
	validEnvs := map[string]bool{
		"SIT": true,
		"UAT": true,
		"PRD": true,
	}
	for _, env := range req.Environments {
		if !validEnvs[env] {
			return fmt.Errorf("invalid environment: %s", env)
		}
	}

	return nil
}

// scanResources scans the template directory for Kubernetes resources
func scanResources(path string) (*ApplicationResources, error) {
	// TODO: Implement resource scanning logic
	// This should walk through the template directory and count different types of resources
	return &ApplicationResources{}, nil
}
