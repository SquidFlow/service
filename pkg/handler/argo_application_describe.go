package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/middleware"
	"github.com/h4-poc/service/pkg/store"
)

type (
	ArgoApplicationDetail struct {
		Name                string              `json:"name"`
		TenantName          string              `json:"tenant_name"`
		AppCode             string              `json:"appcode"`
		Description         string              `json:"description"`
		CreatedBy           string              `json:"created_by"`
		Template            TemplateInfo        `json:"template"`
		DestinationClusters DestinationClusters `json:"destination_clusters"`
		Ingress             []Ingress           `json:"ingress,omitempty"`
		Security            *Security           `json:"security,omitempty"`
		Labels              map[string]string   `json:"labels,omitempty"`
		Annotations         map[string]string   `json:"annotations,omitempty"`
		RuntimeStatus       RuntimeStatusInfo   `json:"runtime_status"`
	}
)

func DescribeArgoApplication(c *gin.Context) {
	tenant := c.GetString(middleware.TenantKey)
	username := c.GetString(middleware.UserNameKey)
	appName := c.Param("name")

	log.G().Infof("tenant: %s, username: %s, appName: %s", tenant, username, appName)

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

	app, err := getApplicationDetail(context.Background(), &AppListOptions{
		CloneOpts:    cloneOpts,
		ProjectName:  tenant,
		ArgoCDClient: argoClient,
	}, appName)

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get application detail: %v", err)})
		return
	}

	c.JSON(200, app)
}

func getApplicationDetail(ctx context.Context, opts *AppListOptions, appName string) (*ArgoApplicationDetail, error) {
	_, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return nil, err
	}

	appPath := repofs.Join(store.Default.AppsDir, appName, store.Default.OverlaysDir, opts.ProjectName)
	log.G().WithFields(log.Fields{
		"AppsDir":     store.Default.AppsDir,
		"OverlaysDir": store.Default.OverlaysDir,
		"project":     opts.ProjectName,
		"appName":     appName,
		"path":        appPath,
	}).Debug("getting application detail")

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

	app := &ArgoApplicationDetail{
		Name:        conf.UserGivenName,
		TenantName:  opts.ProjectName,
		AppCode:     conf.Annotations["h4-poc.github.io/appcode"],
		Description: conf.Annotations["h4-poc.github.io/description"],
		CreatedBy:   conf.Annotations["h4-poc.github.io/created-by"],
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
		Ingress: []Ingress{
			{
				Host: conf.Annotations["h4-poc.github.io/ingress.host"],
				TLS: &TLS{
					Enabled:    conf.Annotations["h4-poc.github.io/ingress.tls.enabled"] == "true",
					SecretName: conf.Annotations["h4-poc.github.io/ingress.tls.secretName"],
				},
			},
		},
		Security: &Security{
			ExternalSecret: &ExternalSecret{
				SecretStoreRef: SecretStoreRef{
					ID: conf.Annotations["h4-poc.github.io/security.external_secret.secret_store_ref.id"],
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

	return app, nil
}
