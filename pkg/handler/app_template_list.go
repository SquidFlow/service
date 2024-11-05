package handler

import (
	"github.com/gin-gonic/gin"
)

// ListTemplateResponse represents the response structure for template listing
type ListTemplateResponse struct {
	Success bool                  `json:"success"`
	Total   int                   `json:"total"`
	Items   []ApplicationTemplate `json:"items"`
}

// ListApplicationTemplate handles the retrieval of application templates
func ListApplicationTemplate(c *gin.Context) {
	// Get query parameters for filtering
	appType := c.Query("appType")     // Filter by application type (helm/kustomize)
	owner := c.Query("owner")         // Filter by owner
	validated := c.Query("validated") // Filter by validation status

	// Get templates from storage with filters
	templates, err := getTemplates(ListTemplateFilter{
		AppType:   appType,
		Owner:     owner,
		Validated: validated,
	})

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
		Total:   len(templates),
		Items:   templates,
	})
}

// ListTemplateFilter represents the filter criteria for listing templates
type ListTemplateFilter struct {
	AppType   string
	Owner     string
	Validated string
}

// getTemplates retrieves templates from storage with optional filters
func getTemplates(filter ListTemplateFilter) ([]ApplicationTemplate, error) {
	// TODO: Implement actual database query
	// This is a mock implementation
	templates := []ApplicationTemplate{
		{
			ID:          1,
			Name:        "h4-loki",
			Path:        "loki",
			Description: "Loki is a horizontally scalable, highly available, multi-tenant log aggregation system inspired by Prometheus.",
			Validated:   true,
			Owner:       "h4-loki",
			AppType:     "helm",
			Source: ApplicationSource{
				Type:   "git",
				URL:    "https://github.com/h4-poc/manifest",
				Branch: "main",
			},
			Environments: []string{"SIT", "UAT", "PRD"},
			LastApplied:  "2024-01-01T00:00:00Z",
			Resources: ApplicationResources{
				Deployments: 1,
				Services:    1,
				Configmaps:  2,
			},
			Events: []ApplicationEvent{
				{
					Time:    "2024-01-01T00:00:00Z",
					Type:    "Normal",
					Reason:  "Created",
					Message: "Application template created successfully",
				},
			},
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		},
		{
			ID:          2,
			Name:        "h4-logging-operator",
			Path:        "logging-operator",
			Description: "Logging operator is a tool for managing logging resources in Kubernetes.",
			Validated:   true,
			Owner:       "h4-logging-operator",
			AppType:     "helm",
			Source: ApplicationSource{
				Type:   "git",
				URL:    "https://github.com/h4-poc/manifest",
				Branch: "main",
			},
			Environments: []string{"SIT", "UAT"},
			LastApplied:  "2024-01-01T00:00:00Z",
			Resources: ApplicationResources{
				Deployments: 1,
				Services:    1,
				Configmaps:  1,
			},
			Events: []ApplicationEvent{
				{
					Time:    "2024-01-01T00:00:00Z",
					Type:    "Normal",
					Reason:  "Created",
					Message: "Application template created successfully",
				},
			},
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		},
	}

	// Apply filters if provided
	var filtered []ApplicationTemplate
	for _, t := range templates {
		if filter.AppType != "" && t.AppType != filter.AppType {
			continue
		}
		if filter.Owner != "" && t.Owner != filter.Owner {
			continue
		}
		if filter.Validated != "" {
			isValidated := filter.Validated == "true"
			if t.Validated != isValidated {
				continue
			}
		}
		filtered = append(filtered, t)
	}

	return filtered, nil
}
