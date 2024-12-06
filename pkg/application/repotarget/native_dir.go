package repotarget

import (
	"context"
	"errors"
	"fmt"

	billyUtils "github.com/go-git/go-billy/v5/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"
	"github.com/squidflow/service/pkg/util"
)

var _ RepoTarget = &NativeRepoTarget{}

// NativeRepoTarget implements the native GitOps repository structure
type NativeRepoTarget struct{}

// RunAppCreate creates an application in the native GitOps repository structure
func (n *NativeRepoTarget) RunAppCreate(ctx context.Context, opts *types.AppCreateOptions) error {
	var (
		appsRepo git.Repository
		appsfs   fs.FS
	)

	log.G().WithFields(log.Fields{
		"app-url":      opts.AppsCloneOpts.URL(),
		"app-revision": opts.AppsCloneOpts.Revision(),
		"app-path":     opts.AppsCloneOpts.Path(),
	}).Debug("starting with options: ")

	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return err
	}
	log.G().Debugf("repofs: %v", repofs)

	if opts.AppsCloneOpts.Repo != "" {
		if opts.AppsCloneOpts.Auth.Password == "" {
			opts.AppsCloneOpts.Auth.Username = opts.CloneOpts.Auth.Username
			opts.AppsCloneOpts.Auth.Password = opts.CloneOpts.Auth.Password
			opts.AppsCloneOpts.Auth.CertFile = opts.CloneOpts.Auth.CertFile
			opts.AppsCloneOpts.Provider = opts.CloneOpts.Provider
		}

		appsRepo, appsfs, err = getRepo(ctx, opts.AppsCloneOpts)
		if err != nil {
			return err
		}
	} else {
		opts.AppsCloneOpts = opts.CloneOpts
		appsRepo, appsfs = r, repofs
	}

	if err = setAppOptsDefaults(ctx, repofs, opts); err != nil {
		return err
	}

	app, err := parseApp(opts.AppOpts, opts.ProjectName, opts.CloneOpts.URL(), opts.CloneOpts.Revision(), opts.CloneOpts.Path())
	if err != nil {
		return fmt.Errorf("failed to parse application from flags: %w", err)
	}

	if err = app.CreateFiles(repofs, appsfs, opts.ProjectName); err != nil {
		if errors.Is(err, application.ErrAppAlreadyInstalledOnProject) {
			return fmt.Errorf("application '%s' already exists in project '%s': %w", app.Name(), opts.ProjectName, err)
		}

		return err
	}

	if opts.AppsCloneOpts != opts.CloneOpts {
		log.G().Info("committing changes to apps repo...")
		if _, err = appsRepo.Persist(ctx, &git.PushOptions{
			CommitMsg: genCommitMsg("chore: "+types.ActionTypeCreate, types.ResourceNameApp, opts.AppOpts.AppName, opts.ProjectName, repofs),
		}); err != nil {
			return fmt.Errorf("failed to push to apps repo: %w", err)
		}
	}

	log.G(ctx).Info("committing changes to gitops repo...")
	revision, err := r.Persist(ctx, &git.PushOptions{
		CommitMsg: genCommitMsg("chore: "+types.ActionTypeCreate, types.ResourceNameApp, opts.AppOpts.AppName, opts.ProjectName, repofs),
	})
	if err != nil {
		return fmt.Errorf("failed to push to gitops repo: %w", err)
	}

	if opts.Timeout > 0 {
		namespace, err := getInstallationNamespace(repofs)
		if err != nil {
			return fmt.Errorf("failed to get application namespace: %w", err)
		}

		log.G(ctx).WithField("timeout", opts.Timeout).Infof("waiting for '%s' to finish syncing", opts.AppOpts.AppName)
		fullName := fmt.Sprintf("%s-%s", opts.ProjectName, opts.AppOpts.AppName)

		// wait for argocd to be ready before applying argocd-apps
		stop := util.WithSpinner(ctx, fmt.Sprintf("waiting for '%s' to be ready", fullName))
		if err = waitAppSynced(ctx, opts.KubeFactory, opts.Timeout, fullName, namespace, revision, true); err != nil {
			stop()
			return fmt.Errorf("failed waiting for application to sync: %w", err)
		}

		stop()
	}

	log.G().Infof("installed application: %s and revision: %s", opts.AppOpts.AppName, revision)
	return nil
}

// RunAppDelete deletes an application from the native GitOps repository structure
func (n *NativeRepoTarget) RunAppDelete(ctx context.Context, opts *types.AppDeleteOptions) error {
	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return err
	}

	appDir := repofs.Join(store.Default.AppsDir, opts.AppName)
	appExists := repofs.ExistsOrDie(appDir)
	if !appExists {
		return fmt.Errorf("application '%s' not found", opts.AppName)
	}

	var dirToRemove string
	commitMsg := fmt.Sprintf("chore: delete app '%s'", opts.AppName)
	if opts.Global {
		dirToRemove = appDir
	} else {
		appOverlaysDir := repofs.Join(appDir, store.Default.OverlaysDir)
		overlaysExists := repofs.ExistsOrDie(appOverlaysDir)
		if !overlaysExists {
			appOverlaysDir = appDir
		}

		appProjectDir := repofs.Join(appOverlaysDir, opts.ProjectName)
		overlayExists := repofs.ExistsOrDie(appProjectDir)
		if !overlayExists {
			return fmt.Errorf("application '%s' not found in project '%s'", opts.AppName, opts.ProjectName)
		}

		allOverlays, err := repofs.ReadDir(appOverlaysDir)
		if err != nil {
			return fmt.Errorf("failed to read overlays directory '%s': %w", appOverlaysDir, err)
		}

		if len(allOverlays) == 1 {
			dirToRemove = appDir
		} else {
			commitMsg += fmt.Sprintf(" from project '%s'", opts.ProjectName)
			dirToRemove = appProjectDir
		}
	}

	err = billyUtils.RemoveAll(repofs, dirToRemove)
	if err != nil {
		return fmt.Errorf("failed to delete directory '%s': %w", dirToRemove, err)
	}

	log.G().Info("committing changes to gitops repo...")
	if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: commitMsg}); err != nil {
		return fmt.Errorf("failed to push to repo: %w", err)
	}

	return nil
}

// ListApps lists all applications in the native GitOps repository structure
func (n *NativeRepoTarget) RunAppList(ctx context.Context, opts *types.AppListOptions) (*types.ApplicationListResponse, error) {
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

	response := &types.ApplicationListResponse{
		Total:        int64(len(matches)),
		Success:      true,
		Applications: make([]types.Application, 0, len(matches)),
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

		app := types.Application{
			ApplicationSource: types.ApplicationSourceRequest{
				Repo:           conf.SrcRepoURL,
				Path:           conf.SrcPath,
				TargetRevision: conf.SrcTargetRevision,
			},
			ApplicationInstantiation: types.ApplicationInstantiation{
				ApplicationName: conf.UserGivenName,
				TenantName:      opts.ProjectName,
				AppCode:         conf.Annotations["squidflow.github.io/appcode"],
				Description:     conf.Annotations["squidflow.github.io/description"],
			},
			ApplicationTarget: []types.ApplicationTarget{
				{
					Cluster:   "in-cluster",
					Namespace: conf.DestNamespace,
				},
			},
			ApplicationRuntime: types.ApplicationRuntime{
				GitInfo:         []types.GitInfo{*gitInfo},
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

// RunAppUpdate updates an application in the native GitOps repository structure
func (n *NativeRepoTarget) RunAppUpdate(ctx context.Context, opts *types.UpdateOptions) error {
	return nil
}

// RunAppGet gets an application from the native GitOps repository structure
func (n *NativeRepoTarget) RunAppGet(ctx context.Context, opts *types.AppListOptions, appName string) (*types.Application, error) {
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

	app := &types.Application{
		ApplicationSource: types.ApplicationSourceRequest{
			Repo:           conf.SrcRepoURL,
			Path:           conf.SrcPath,
			TargetRevision: conf.SrcTargetRevision,
		},
		ApplicationInstantiation: types.ApplicationInstantiation{
			ApplicationName: conf.AppName,
			TenantName:      opts.ProjectName,
			AppCode:         conf.Annotations["squidflow.github.io/appcode"],
			Description:     conf.Annotations["squidflow.github.io/description"],
		},
		ApplicationTarget: []types.ApplicationTarget{
			{
				Cluster:   "default",
				Namespace: conf.DestNamespace,
			},
		},
		ApplicationRuntime: types.ApplicationRuntime{
			GitInfo:         []types.GitInfo{*gitInfo},
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
