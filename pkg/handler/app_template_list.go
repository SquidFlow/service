package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/spf13/viper"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

// ListTemplateResponse represents the response structure for template listing
type ListTemplateResponse struct {
	Success bool                  `json:"success"`
	Total   int                   `json:"total"`
	Items   []ApplicationTemplate `json:"items"`
	Message string                `json:"message"`
}

// ListApplicationTemplate handles the retrieval of application templates
func ListApplicationTemplate(c *gin.Context) {
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

	templates, err := RunApplicationTemplateList(context.Background(), &AppTemplateListOptions{
		CloneOpts: cloneOpts,
	})
	if err != nil {
		log.G().Errorf("Failed to list application templates: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list application templates: %v", err)})
		return
	}

	// Return success response
	c.JSON(200, ListTemplateResponse{
		Success: true,
		Total:   len(templates),
		Items:   templates,
		Message: "success",
	})
}

// RunApplicationTemplateList retrieves the application templates from the gitops repository
func RunApplicationTemplateList(ctx context.Context, opts *AppTemplateListOptions) ([]ApplicationTemplate, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return nil, err
	}

	// Look for template files in the cluster resources directory
	matches, err := billyUtils.Glob(repofs, repofs.Join(
		store.Default.BootsrtrapDir,
		store.Default.ClusterResourcesDir,
		store.Default.ClusterContextName,
		"apptemp-*.yaml",
	))
	if err != nil {
		return nil, err
	}

	var templates []ApplicationTemplate

	for _, file := range matches {
		log.G().WithField("file", file).Debug("Found application template")

		template := &apptempv1alpha1.ApplicationTemplate{}
		if err := repofs.ReadYamls(file, template); err != nil {
			log.G().Warnf("Failed to read template from %s: %v", file, err)
			continue
		}

		log.G().WithFields(log.Fields{
			"id":          template.Annotations["h4-poc.github.io/id"],
			"name":        template.Name,
			"owner":       template.Annotations["h4-poc.github.io/owner"],
			"description": template.Annotations["h4-poc.github.io/description"],
			"created-at":  template.Annotations["h4-poc.github.io/created-at"],
			"updated-at":  template.Annotations["h4-poc.github.io/updated-at"],
		}).Debug("Found application template")

		// Convert to response type
		appTemplate := ApplicationTemplate{
			ID:          template.Annotations["h4-poc.github.io/id"],
			Name:        template.Name,
			Owner:       template.Annotations["h4-poc.github.io/owner"],
			Description: template.Annotations["h4-poc.github.io/description"],
			AppType:     getAppTempType(*template),
			Validated:   true,
			Path:        template.Spec.Helm.RenderTargets[0].ValuesPath,
			CreatedAt:   template.Annotations["h4-poc.github.io/created-at"],
			UpdatedAt:   template.Annotations["h4-poc.github.io/updated-at"],
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
		}
		templates = append(templates, appTemplate)
	}

	return templates, nil
}
