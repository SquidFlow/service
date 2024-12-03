package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
)

func DescribeApplicationHandler(c *gin.Context) {
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

func getApplicationDetail(ctx context.Context, opts *AppListOptions, appName string) (*Application, error) {
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

	app := &Application{
		ApplicationSource: ApplicationSource{
			Repo:           conf.SrcRepoURL,
			Path:           conf.SrcPath,
			TargetRevision: conf.SrcTargetRevision,
		},
		ApplicationInstantiation: ApplicationInstantiation{
			ApplicationName: conf.AppName,
			TenantName:      opts.ProjectName,
			AppCode:         conf.Annotations["squidflow.github.io/appcode"],
			Description:     conf.Annotations["squidflow.github.io/description"],
		},
		ApplicationTarget: []ApplicationTarget{
			{
				Cluster:   "default",
				Namespace: conf.DestNamespace,
			},
		},
		ApplicationRuntime: ApplicationRuntime{
			GitInfo:         []GitInfo{*gitInfo},
			ResourceMetrics: *resourceMetrics,
		},
	}

	argoApp, err := opts.ArgoCDClient.Applications(applicationNs).Get(ctx, applicationName, metav1.GetOptions{})
	if err != nil {
		log.G().Warnf("failed to get ArgoCD app info for %s: %v", conf.UserGivenName, err)
	}

	app.ApplicationRuntime.Status = getAppStatus(argoApp)
	app.ApplicationRuntime.Health = getAppHealth(argoApp)
	app.ApplicationRuntime.SyncStatus = getAppSyncStatus(argoApp)

	return app, nil
}
