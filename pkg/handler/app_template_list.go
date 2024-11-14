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
}

// ListTemplateFilter represents the filter criteria for listing templates
type ListTemplateFilter struct {
	AppType   string
	Owner     string
	Validated string
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
		"*.yaml",
	))
	if err != nil {
		return nil, err
	}

	var templates []ApplicationTemplate

	for _, name := range matches {
		log.G().Debugf("Reading template from %s", name)
		template := &apptempv1alpha1.ApplicationTemplate{}
		if err := repofs.ReadYamls(name, template); err != nil {
			log.G().Warnf("Failed to read template from %s: %v", name, err)
			continue
		}

		if template.Kind != "ApplicationTemplate" {
			log.G().Warnf("skip %s/%s", template.Kind, name)
			continue
		}

		// Convert to response type
		appTemplate := ApplicationTemplate{
			Name:        template.Name,
			Owner:       template.Annotations["owner"],
			Description: template.Annotations["description"],
			AppType:     getAppTempType(*template),
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

func getAppTempType(temp apptempv1alpha1.ApplicationTemplate) ApplicationTemplateType {
	var enableHelm, enableKustomize bool
	if temp.Spec.Helm != nil {
		enableHelm = true
	}
	if temp.Spec.Kustomize != nil {
		enableKustomize = true
	}

	// only define 2 types: helm and kustomize
	if enableHelm && enableKustomize {
		return ApplicationTemplateTypeHelmKustomize
	}
	if enableHelm {
		return ApplicationTemplateTypeHelm
	}
	if enableKustomize {
		return ApplicationTemplateTypeKustomize
	}
	return "unknown"
}
