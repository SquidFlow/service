package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

// GetTemplateResponse represents the response structure for getting a single template
type GetTemplateResponse struct {
	Item    ApplicationTemplate `json:"item"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
}

// DescribeApplicationTemplate handles the retrieval of a single application template
func DescribeApplicationTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		c.JSON(400, gin.H{"error": "template id is required"})
		return
	}

	// Get kube factory from context
	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: false,
	}
	cloneOpts.Parse()

	template, err := GetApplicationTemplateFromRepo(context.Background(), &AppTemplateGetOptions{
		CloneOpts: cloneOpts,
		Name:      fmt.Sprintf("apptemp-%s.yaml", templateID),
	}, templateID)
	if err != nil {
		log.G().Errorf("Failed to get application template: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get application template: %v", err)})
		return
	}

	if template == nil {
		c.JSON(404, gin.H{"error": "template not found"})
		return
	}

	// Return success response
	c.JSON(200, GetTemplateResponse{
		Success: true,
		Item:    *template,
	})
}

// AppTemplateGetOptions contains options for getting a single application template
type AppTemplateGetOptions struct {
	CloneOpts *git.CloneOptions
	Name      string
}

// GetApplicationTemplateFromRepo retrieves a single application template from the gitops repository
func GetApplicationTemplateFromRepo(ctx context.Context, opts *AppTemplateGetOptions, id string) (*ApplicationTemplate, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return nil, err
	}

	templatePath := repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		fmt.Sprintf("apptemp-%s.yaml", id),
	)

	template := &apptempv1alpha1.ApplicationTemplate{}
	if err := repofs.ReadYamls(templatePath, template); err != nil {
		return nil, fmt.Errorf("failed to read template %s: %v", opts.Name, err)
	}

	if template.Kind != "ApplicationTemplate" {
		return nil, fmt.Errorf("invalid template kind: %s", template.Kind)
	}

	return &ApplicationTemplate{
		Name:        template.Name,
		Owner:       template.Annotations["h4-poc.github.io/owner"],
		Description: template.Annotations["h4-poc.github.io/description"],
		ID:          template.Annotations["h4-poc.github.io/id"],
		CreatedAt:   template.Annotations["h4-poc.github.io/created-at"],
		UpdatedAt:   template.Annotations["h4-poc.github.io/updated-at"],
		AppType:     getAppTempType(*template),
		Validated:   true,
		Path:        template.Spec.Helm.RenderTargets[0].ValuesPath,
		Source: ApplicationSource{
			URL:            template.Spec.RepoURL,
			TargetRevision: template.Spec.TargetRevision,
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
	}, nil
}
