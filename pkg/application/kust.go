package application

import (
	"fmt"
	"path"

	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	kusttypes "sigs.k8s.io/kustomize/api/types"

	"github.com/squidflow/service/pkg/application/dryrun"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/util"
)

type kustApp struct {
	baseApp
	base      *kusttypes.Kustomization
	overlay   *kusttypes.Kustomization
	manifests []byte
	namespace *v1.Namespace
	config    *Config
}

/* kustApp Application impl */
func newKustApp(o *CreateOptions, projectName, repoURL, targetRevision, repoRoot string) (*kustApp, error) {
	var err error
	app := &kustApp{
		baseApp: baseApp{o},
	}

	if o.AppSpecifier == "" {
		return nil, ErrEmptyAppSpecifier
	}

	if o.AppName == "" {
		return nil, ErrEmptyAppName
	}

	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	app.base = &kusttypes.Kustomization{
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
		Resources: []string{o.AppSpecifier},
	}

	if o.InstallationMode != "" && !o.InstallationMode.isValid() {
		return nil, fmt.Errorf("unknown installation mode: %s", o.InstallationMode)
	}

	if o.InstallationMode == InstallModeFlatten {
		host, orgRepo, path, gitRef, _, _, _ := util.ParseGitUrl(o.AppSpecifier)
		log.G().WithFields(log.Fields{
			"host":    host,
			"orgRepo": orgRepo,
			"path":    path,
			"gitRef":  gitRef,
		}).Debug("parsed git url, kustomizing manifests")

		_, appfs, exists := git.GetRepositoryCache().Get(orgRepo, false)
		if !exists {
			return nil, fmt.Errorf("repo not found in cache")
		}

		app.manifests, err = dryrun.GenerateKustomizeManifest(appfs, path, "default")
		if err != nil {
			return nil, err
		}

		app.base.Resources[0] = "manifest.yaml"
	}

	app.overlay = &kusttypes.Kustomization{
		Resources: []string{
			"../../base",
		},
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
	}

	if o.DestNamespace != "" && o.DestNamespace != "default" {
		app.overlay.Namespace = o.DestNamespace
		app.namespace = kube.GenerateNamespace(o.DestNamespace, nil)
	}

	log.G().WithFields(log.Fields{
		"appName":           o.AppName,
		"destNamespace":     o.DestNamespace,
		"destServer":        o.DestServer,
		"srcRepoURL":        repoURL,
		"srcPath":           path.Join(repoRoot, store.Default.AppsDir, o.AppName, store.Default.OverlaysDir, projectName),
		"srcTargetRevision": targetRevision,
		"labels":            o.Labels,
		"annotations":       o.Annotations,
		"installationMode":  o.InstallationMode,
	}).Debug("creating app config")

	app.config = &Config{
		AppName:           o.AppName,
		UserGivenName:     o.AppName,
		DestNamespace:     o.DestNamespace,
		DestServer:        o.DestServer,
		SrcRepoURL:        repoURL,
		SrcPath:           path.Join(repoRoot, store.Default.AppsDir, o.AppName, store.Default.OverlaysDir, projectName),
		SrcTargetRevision: targetRevision,
		Labels:            o.Labels,
		Annotations:       o.Annotations,
	}

	return app, nil
}

func (app *kustApp) CreateFiles(repofs fs.FS, appsfs fs.FS, projectName string) error {
	return kustCreateFiles(app, repofs, appsfs, projectName)
}

func (app *kustApp) Manifests() map[string][]byte {
	return map[string][]byte{
		"default": app.manifests,
	}
}

func kustCreateFiles(app *kustApp, repofs fs.FS, appsfs fs.FS, projectName string) error {
	var err error

	// create Base
	appPath := appsfs.Join(store.Default.AppsDir, app.Name())
	basePath := appsfs.Join(appPath, "base")
	baseKustomizationPath := appsfs.Join(basePath, "kustomization.yaml")

	// check if app is in the same filesystem
	if appsfs.ExistsOrDie(appPath) {
		// check if the bases are the same
		log.G().Debug("application with the same name exists, checking for collisions")
		if collision, err := checkBaseCollision(appsfs, baseKustomizationPath, app.base); err != nil {
			return err
		} else if collision {
			return ErrAppCollisionWithExistingBase
		}
	} else if appsfs != repofs && repofs.ExistsOrDie(appPath) {
		appRepo, err := getAppRepo(repofs, app.Name())
		if err != nil {
			return fmt.Errorf("Failed getting app repo: %w", err)
		}

		return fmt.Errorf("an application with the same name already exists in '%s', consider choosing a different name", appRepo)
	}

	if err = appsfs.WriteYamls(baseKustomizationPath, app.base); err != nil {
		return err
	}

	// create Overlay
	overlayPath := appsfs.Join(appPath, "overlays", projectName)
	overlayKustomizationPath := appsfs.Join(overlayPath, "kustomization.yaml")
	log.G().Debugf("overlay path: %s", overlayKustomizationPath)
	if appsfs.ExistsOrDie(overlayKustomizationPath) {
		return ErrAppAlreadyInstalledOnProject
	}

	if err = appsfs.WriteYamls(overlayKustomizationPath, app.overlay); err != nil {
		return err
	}

	// create manifests - only used in flat installation mode
	if app.manifests != nil {
		manifestsPath := appsfs.Join(basePath, "manifest.yaml")
		if _, err = writeFile(appsfs, manifestsPath, "manifests", app.manifests); err != nil {
			return err
		}
	}

	clusterName, err := getClusterName(repofs, app.opts.DestServer)
	if err != nil {
		return err
	}

	if app.namespace != nil {
		if err = createNamespaceManifest(repofs, clusterName, app.namespace); err != nil {
			return err
		}
	}

	configPath := repofs.Join(overlayPath, "config.json")
	if repofs != appsfs {
		configPath = repofs.Join(appPath, projectName, "config.json")
	}

	if viper.GetString("gitops.mode") == "pull_request" {
		if err = appsfs.WriteJson(configPath, app.config); err != nil {
			return fmt.Errorf("appsfs failed to write app config.json: %w", err)
		}
		return nil
	}

	if err = repofs.WriteJson(configPath, app.config); err != nil {
		return fmt.Errorf("repofs failed to write app config.json: %w", err)
	}

	return nil
}
