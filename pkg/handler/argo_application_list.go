package handler

import (
	"context"
	"fmt"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argocdv1alpha1client "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5"
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

type ResourceMetrics struct {
	CPUCores    string `json:"cpu_cores"`
	MemoryUsage string `json:"memory_usage"`
}

type (
	ArgoApplicationListResponse struct {
		Success bool                    `json:"success"`
		Total   int                     `json:"total"`
		Items   []ArgoApplicationDetail `json:"items"`
	}

	ArgoApplicationDetail struct {
		Name                string              `json:"name"`
		TenantName          string              `json:"tenant_name"`
		AppCode             string              `json:"appcode"`
		Description         string              `json:"description"`
		CreatedBy           string              `json:"created_by"`
		Template            TemplateInfo        `json:"template"`
		DestinationClusters DestinationClusters `json:"destination_clusters"`
		Ingress             *Ingress            `json:"ingress,omitempty"`
		Security            *Security           `json:"security,omitempty"`
		Labels              map[string]string   `json:"labels,omitempty"`
		Annotations         map[string]string   `json:"annotations,omitempty"`
		RuntimeStatus       RuntimeStatusInfo   `json:"runtime_status"`
	}

	TemplateInfo struct {
		Source         ApplicationSource `json:"source"`
		LastCommitInfo GitInfo           `json:"last_commit_info"`
	}

	RuntimeStatusInfo struct {
		Status           string          `json:"status"`
		Health           string          `json:"health"`
		SyncStatus       string          `json:"sync_status"`
		DeployedClusters []ClusterStatus `json:"deployed_clusters"`
		ResourceMetrics  ResourceMetrics `json:"resource_metrics"`
	}

	ClusterStatus struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		PodCount     int    `json:"pod_count"`
		SecretCount  int    `json:"secret_count"`
		Status       string `json:"status"`
		LastSyncTime string `json:"last_sync_time"`
	}
)

func ListArgoApplications(c *gin.Context) {
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

func RunAppList(ctx context.Context, opts *AppListOptions) (*ArgoApplicationListResponse, error) {
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

	response := &ArgoApplicationListResponse{
		Total: len(matches),
		Items: make([]ArgoApplicationDetail, 0, len(matches)),
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

		app := ArgoApplicationDetail{
			Name:        conf.UserGivenName,
			TenantName:  opts.ProjectName,
			AppCode:     conf.Annotations["squidflow.github.io/appcode"],
			Description: conf.Annotations["squidflow.github.io/description"],
			CreatedBy:   conf.Annotations["squidflow.github.io/created-by"],
			Template: TemplateInfo{
				Source: ApplicationSource{
					Type: string(ApplicationTemplateTypeKustomize),
					Path: conf.SrcPath,
					URL:  conf.SrcRepoURL,
				},
				LastCommitInfo: GitInfo{
					LastCommitID:      gitInfo.LastCommitID,
					LastCommitMessage: gitInfo.LastCommitMessage,
				},
			},
			DestinationClusters: DestinationClusters{
				Clusters:  []string{"in-cluster"},
				Namespace: conf.DestNamespace,
			},
			Ingress: &Ingress{
				Host: conf.Annotations["squidflow.github.io/ingress.host"],
				TLS: &TLS{
					Enabled:    conf.Annotations["squidflow.github.io/ingress.tls.enabled"] == "true",
					SecretName: conf.Annotations["squidflow.github.io/ingress.tls.secretName"],
				},
			},
			Security: &Security{
				ExternalSecret: &ExternalSecret{
						SecretStoreRef: SecretStoreRef{
							ID: conf.Annotations["squidflow.github.io/security.external_secret.secret_store_ref.id"],
						},
					},
			},
			RuntimeStatus: RuntimeStatusInfo{
				ResourceMetrics: ResourceMetrics{
					CPUCores:    resourceMetrics.CPU,
					MemoryUsage: resourceMetrics.Memory,
				},
			},
		}

		// runtime status
		argoApp, err := opts.ArgoCDClient.Applications(applicationNs).Get(ctx, applicationName, metav1.GetOptions{})
		if err != nil {
			log.G().Warnf("failed to get ArgoCD app info for %s: %v", conf.UserGivenName, err)
			app.RuntimeStatus.Status = "Unknown"
			app.RuntimeStatus.Health = "Unknown"
			app.RuntimeStatus.SyncStatus = "Unknown"
		} else {
			app.RuntimeStatus.Status = getAppStatus(argoApp)
			app.RuntimeStatus.Health = getAppHealth(argoApp)
			app.RuntimeStatus.SyncStatus = getAppSyncStatus(argoApp)
		}

		response.Items = append(response.Items, app)
	}

	return response, nil
}

type GitInfo struct {
	Creator           string
	LastUpdater       string
	LastCommitID      string
	LastCommitMessage string
}

// TODO: Implement this function later
func getGitInfo(repofs billy.Filesystem, appPath string) (*GitInfo, error) {
	return &GitInfo{
		Creator:           "Unknown",
		LastUpdater:       "Unknown",
		LastCommitID:      "Unknown",
		LastCommitMessage: "Unknown",
	}, nil
}

type ResourceMetricsInfo struct {
	PodCount    int
	SecretCount int
	CPU         string
	Memory      string
}

// TODO: Implement this function later
func getResourceMetrics(ctx context.Context, kubeClient kubernetes.Interface, namespace string) (*ResourceMetricsInfo, error) {
	return &ResourceMetricsInfo{
		PodCount:    5,
		SecretCount: 12,
		CPU:         "0.25",
		Memory:      "200Mi",
	}, nil
}

func getAppStatus(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}
	log.G().Debugf("get app OperationState: %v", app.Status.OperationState)
	return string(app.Status.OperationState.Phase)
}

func getAppHealth(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}
	log.G().Debugf("get app Health: %v", app.Status.Health.Status)
	return string(app.Status.Health.Status)
}

func getAppSyncStatus(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}
	return string(app.Status.Sync.Status)
}
