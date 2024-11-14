package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"
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
	Item    ApplicationTemplate `json:"item"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
}

// CreateApplicationTemplate handles HTTP requests to create an ApplicationTemplate
func CreateApplicationTemplate(c *gin.Context) {
	var req CreateApplicationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	argocdArgoTemplate, err := generateApplicationTemplate(context.Background(), &req)
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

	err = writeAppTemplate2Repo(
		context.Background(),
		argocdArgoTemplate,
		&AppTemplateCreateOptions{
			CloneOpts: cloneOpts,
		})

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create application template: %v", err)})
		return
	}

	c.JSON(201, CreateApplicationTemplateResponse{
		Item: ApplicationTemplate{
			Name:        argocdArgoTemplate.Name,
			Owner:       argocdArgoTemplate.Annotations["h4-poc.github.io/owner"],
			Description: argocdArgoTemplate.Annotations["h4-poc.github.io/description"],
			ID:          argocdArgoTemplate.Annotations["h4-poc.github.io/id"],
			CreatedAt:   argocdArgoTemplate.Annotations["h4-poc.github.io/created-at"],
			UpdatedAt:   argocdArgoTemplate.Annotations["h4-poc.github.io/updated-at"],
			AppType:     getAppTempType(*argocdArgoTemplate),
			Validated:   true,
			Path:        argocdArgoTemplate.Spec.Helm.RenderTargets[0].ValuesPath,
			Source: ApplicationSource{
				URL:            argocdArgoTemplate.Spec.RepoURL,
				TargetRevision: argocdArgoTemplate.Spec.TargetRevision,
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
		},
		Success: true,
		Message: "application template created",
	})
}

func generateApplicationTemplate(ctx context.Context, userReq *CreateApplicationTemplateRequest) (*apptempv1alpha1.ApplicationTemplate, error) {
	log.G().WithField("user input request", userReq).Info("create application template...")

	template := &apptempv1alpha1.ApplicationTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "argocd-addon.github.com/v1alpha1",
			Kind:       "ApplicationTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      userReq.Name,
			Namespace: store.Default.ArgoCDNamespace,
			Annotations: map[string]string{
				"h4-poc.github.io/id":          getNewId(),
				"h4-poc.github.io/owner":       userReq.Owner,
				"h4-poc.github.io/description": userReq.Description,
				"h4-poc.github.io/created-at":  time.Now().Format(time.RFC3339),
				"h4-poc.github.io/updated-at":  time.Now().Format(time.RFC3339),
			},
		},
		Spec: apptempv1alpha1.ApplicationTemplateSpec{
			Name:           userReq.Name,
			RepoURL:        userReq.Source.URL,
			TargetRevision: userReq.Source.TargetRevision,
			Helm: &apptempv1alpha1.HelmConfig{
				Chart:      userReq.Name,
				Version:    "v1",
				Repository: userReq.Source.URL,
				RenderTargets: []apptempv1alpha1.HelmRenderTarget{
					{
						DestinationCluster: apptempv1alpha1.ClusterSelector{
							Name: userReq.Owner,
							MatchLabels: map[string]string{
								"env": userReq.Owner,
							},
						},
						ValuesPath: userReq.Path,
					},
				},
			},
			Kustomize: &apptempv1alpha1.KustomizeConfig{
				RenderTargets: []apptempv1alpha1.KustomizeRenderTarget{
					{
						Path: userReq.Path,
						DestinationCluster: apptempv1alpha1.ClusterSelector{
							Name: userReq.Owner,
							MatchLabels: map[string]string{
								"env": userReq.Owner,
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

func writeAppTemplate2Repo(ctx context.Context, appTemplate *apptempv1alpha1.ApplicationTemplate, opts *AppTemplateCreateOptions) error {
	var (
		err error
	)

	log.G().WithField("cloneOpts", opts.CloneOpts).Debug("run app template create with clone options")
	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return err
	}

	appTemplateYaml, err := yaml.Marshal(appTemplate)
	if err != nil {
		return err
	}

	appTemplateExists := repofs.ExistsOrDie(
		repofs.Join(store.Default.BootsrtrapDir,
			store.Default.ClusterResourcesDir,
			store.Default.ClusterContextName,
			"apptemp"+appTemplate.Annotations["h4-poc.github.io/id"]+".yaml"),
	)
	if appTemplateExists {
		return fmt.Errorf("app template '%s' already exists", appTemplate.GetName())
	}
	log.G().Debug("repository is ok")

	bulkWrites := []fs.BulkWriteRequest{}
	bulkWrites = append(bulkWrites, fs.BulkWriteRequest{
		Filename: repofs.Join(
			store.Default.BootsrtrapDir,
			store.Default.ClusterResourcesDir,
			store.Default.ClusterContextName,
			fmt.Sprintf("apptemp-%s.yaml", appTemplate.Annotations["h4-poc.github.io/id"]),
		),
		Data:   util.JoinManifests(appTemplateYaml),
		ErrMsg: "failed to create app template file",
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
