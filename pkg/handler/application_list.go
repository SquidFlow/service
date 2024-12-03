package handler

import (
	"context"
	"fmt"

	argocdv1alpha1client "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
)

type AppListOptions struct {
	CloneOpts    *git.CloneOptions
	ProjectName  string
	KubeClient   kubernetes.Interface
	ArgoCDClient *argocdv1alpha1client.ArgoprojV1alpha1Client
}

func ListApplicationsHandler(c *gin.Context) {
	tenant := c.GetString(middleware.TenantKey)
	username := c.GetString(middleware.UserNameKey)

	log.G().Infof("tenant: %s, username: %s", tenant, username)

	var project = tenant

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

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	apps, err := RunAppList(context.Background(), &AppListOptions{
		CloneOpts:    cloneOpts,
		ProjectName:  project,
		ArgoCDClient: argoClient,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list applications: %v", err)})
		return
	}

	c.JSON(200, apps)
}

func RunAppList(ctx context.Context, opts *AppListOptions) (*ApplicationListResponse, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return nil, err
	}

	path := repofs.Join(store.Default.AppsDir, "*", store.Default.OverlaysDir, opts.ProjectName)
	log.G().WithFields(log.Fields{
		"AppsDir":     store.Default.AppsDir,
		"OverlaysDir": store.Default.OverlaysDir,
		"project":     opts.ProjectName,
		"path":        path,
	}).Debug("listing applications")

	matches, err := billyUtils.Glob(repofs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to run glob on %s: %w", opts.ProjectName, err)
	}

	response := &ApplicationListResponse{
		Total:        int64(len(matches)),
		Success:      true,
		Applications: make([]Application, 0, len(matches)),
	}

	for _, appPath := range matches {
		conf, err := getConfigFileFromPath(repofs, appPath)
		if err != nil {
			return nil, err
		}

		gitInfo, err := getGitInfo(repofs, appPath)
		if err != nil {
			log.G().Warnf("failed to get git info for %s: %v", appPath, err)
		}

		var (
			applicationName = opts.ProjectName + "-" + conf.UserGivenName
			applicationNs   = store.Default.ArgoCDNamespace
		)
		log.G().Debugf("applicationName: %s, applicationNs: %s", applicationName, applicationNs)

		resourceMetrics, err := getResourceMetrics(ctx, opts.KubeClient, conf.DestNamespace)
		if err != nil {
			log.G().Warnf("failed to get resource metrics for %s: %v", conf.DestNamespace, err)
		}

		app := Application{
			ApplicationSource: ApplicationSource{
				Repo:           conf.SrcRepoURL,
				Path:           conf.SrcPath,
				TargetRevision: conf.SrcTargetRevision,
			},
			ApplicationInstantiation: ApplicationInstantiation{
				ApplicationName: conf.UserGivenName,
				TenantName:      opts.ProjectName,
				AppCode:         conf.Annotations["squidflow.github.io/appcode"],
				Description:     conf.Annotations["squidflow.github.io/description"],
			},
			ApplicationTarget: []ApplicationTarget{
				{
					Cluster:   "in-cluster",
					Namespace: conf.DestNamespace,
				},
			},
			ApplicationRuntime: ApplicationRuntime{
				GitInfo:         []GitInfo{*gitInfo},
				ResourceMetrics: *resourceMetrics,
			},
		}

		// Get runtime status from ArgoCD
		argoApp, err := opts.ArgoCDClient.Applications(applicationNs).Get(ctx, applicationName, metav1.GetOptions{})
		if err != nil {
			log.G().Warnf("failed to get ArgoCD app info for %s: %v", conf.UserGivenName, err)
			app.ApplicationRuntime.Status = "Unknown"
			app.ApplicationRuntime.Health = "Unknown"
			app.ApplicationRuntime.SyncStatus = "Unknown"
		} else {
			app.ApplicationRuntime.Status = getAppStatus(argoApp)
			app.ApplicationRuntime.Health = getAppHealth(argoApp)
			app.ApplicationRuntime.SyncStatus = getAppSyncStatus(argoApp)
		}

		response.Applications = append(response.Applications, app)
	}

	return response, nil
}
