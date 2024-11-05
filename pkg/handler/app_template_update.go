package handler

import (
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateTemplateRequest represents the request for updating an application template
type UpdateTemplateRequest struct {
	Name        string            `json:"name" binding:"required"`
	Path        string            `json:"path" binding:"required"`
	Owner       string            `json:"owner" binding:"required"`
	Source      ApplicationSource `json:"source" binding:"required"`
	Description string            `json:"description,omitempty"`
	AppType     string            `json:"appType" binding:"required,oneof=kustomize helm"`
}

// UpdateTemplateResponse represents the response for template update operation
type UpdateTemplateResponse struct {
	Success bool   `json:"success"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
}

// TODO: UpdateApplicationTemplate handles the update of an existing application template
func UpdateApplicationTemplate(c *gin.Context) {
	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, UpdateTemplateResponse{
			Success: false,
			ID:      0,
			Name:    req.Name,
		})
		return
	}

	// Generate template ID (mock implementation)
	templateID := 1

	// Create new template with minimal required fields
	template := ApplicationTemplate{
		ID:          templateID,
		Name:        req.Name,
		Path:        req.Path,
		Owner:       req.Owner,
		AppType:     req.AppType,
		Source:      req.Source,
		Description: req.Description,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Events: []ApplicationEvent{
			{
				Time:    time.Now().UTC().Format(time.RFC3339),
				Type:    "Normal",
				Reason:  "Created",
				Message: "Application template created successfully",
			},
		},
	}

	// Return success response with mock data
	c.JSON(201, UpdateTemplateResponse{
		Success: true,
		ID:      template.ID,
		Name:    template.Name,
	})
}
