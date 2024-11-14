package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
)

// CreateApplicationTemplateRequest defines the request body for creating an ApplicationTemplate
type CreateApplicationTemplateRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Path        string                  `json:"path" binding:"required"`
	Owner       string                  `json:"owner" binding:"required"`
	Source      ApplicationSource       `json:"source" binding:"required"`
	Description string                  `json:"description,omitempty"`
	AppType     ApplicationTemplateType `json:"appType" binding:"required,oneof=kustomize helm helm+kustomize"`
}

type CreateApplicationTemplateResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

// CreateApplicationTemplate handles HTTP requests to create an ApplicationTemplate
func CreateApplicationTemplate(c *gin.Context) {
	dynamicClient := c.MustGet("dynamicClient").(dynamic.Interface)
	discoveryClient := c.MustGet("discoveryClient").(*discovery.DiscoveryClient)

	var req CreateApplicationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	temp, err := generateApplicationTemplate(context.Background(), dynamicClient, discoveryClient, &req)
	if err != nil {
		if k8serror.IsAlreadyExists(err) {
			c.JSON(409, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create application template: %v", err)})
		return
	}

	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: true,
	}
	cloneOpts.Parse()

	err = RunAppTemplateCreate(
		context.Background(),
		temp,
		&AppTemplateCreateOptions{
			CloneOpts: cloneOpts,
		})

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create application template: %v", err)})
		return
	}

	c.JSON(201, CreateApplicationTemplateResponse{
		ID:      1,
		Name:    temp.GetName(),
		Success: true,
	})
}

func generateApplicationTemplate(ctx context.Context, dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, userReq *CreateApplicationTemplateRequest) (*unstructured.Unstructured, error) {
	log.G().WithField("user input request", userReq).Info("create application template...")

	template := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argocd-addon.github.com/v1alpha1",
			"kind":       "ApplicationTemplate",
			"metadata": map[string]interface{}{
				"name":      userReq.Name,
				"namespace": store.Default.ArgoCDNamespace,
			},
			"spec": map[string]interface{}{
				"name":           userReq.Name,
				"repoURL":        userReq.Source.URL,
				"targetRevision": userReq.Source.TargetRevision,
				"helm": map[string]interface{}{
					"chart":      userReq.Name,
					"version":    "v1",
					"repository": userReq.Source.URL,
					"renderTargets": []map[string]interface{}{
						{
							"destinationCluster": map[string]interface{}{
								"name": userReq.Owner,
								"matchLabels": map[string]interface{}{
									"env": userReq.Owner,
								},
							},
							"valuesPath": userReq.Path,
						},
					},
				},
				"kustomize": map[string]interface{}{
					"renderTargets": []map[string]interface{}{
						{
							"path": userReq.Path,
							"destinationCluster": map[string]interface{}{
								"name": userReq.Owner,
								"matchLabels": map[string]interface{}{
									"env": userReq.Owner,
								},
							},
						},
					},
				},
			},
		},
	}

	log.G().WithField("template", template).Debug("created application template")

	return template, nil
}

func RunAppTemplateCreate(ctx context.Context, appTemplate *unstructured.Unstructured, opts *AppTemplateCreateOptions) error {
	var (
		err error
	)

	log.G().WithField("cloneOpts", opts.CloneOpts).Debug("clone options")

	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return err
	}

	// convert the appTemplate to yaml
	appTemplateYaml, err := yaml.Marshal(appTemplate)
	if err != nil {
		return err
	}

	appTemplateExists := repofs.ExistsOrDie(
		repofs.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, store.Default.ClusterContextName, appTemplate.GetName()+".yaml"),
	)
	if appTemplateExists {
		return fmt.Errorf("app template '%s' already exists", appTemplate.GetName())
	}
	log.G().Debug("repository is ok")

	bulkWrites := []fs.BulkWriteRequest{}
	bulkWrites = append(bulkWrites, fs.BulkWriteRequest{
		Filename: repofs.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, store.Default.ClusterContextName, appTemplate.GetName()+".yaml"),
		Data:     util.JoinManifests(appTemplateYaml),
		ErrMsg:   "failed to create app template file",
	})

	if err = fs.BulkWrite(repofs, bulkWrites...); err != nil {
		log.G().Errorf("failed to write new app template manifest to repo: %w", err)
		return err
	}

	log.G().Infof("pushing new app template manifest to repo")
	if _, err = r.Persist(ctx,
		&git.PushOptions{
			CommitMsg: fmt.Sprintf("chore: added app template '%s'", appTemplate.GetName()),
		}); err != nil {
		log.G().Errorf("failed to push new app template manifest to repo: %w", err)
		return err
	}

	log.G().Infof("app template created: '%s'", appTemplate.GetName())

	return nil
}
