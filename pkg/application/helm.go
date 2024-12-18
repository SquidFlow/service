package application

import (
	"fmt"
	"path"

	v1 "k8s.io/api/core/v1"
	kusttypes "sigs.k8s.io/kustomize/api/types"

	"github.com/spf13/viper"
	"github.com/squidflow/service/pkg/application/dryrun"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/util"
)

// for helm app, we generate manifests for each environment
// the manifests are stored in the app.manifests map
type helmApp struct {
	baseApp
	base      *kusttypes.Kustomization
	overlay   *kusttypes.Kustomization
	manifests []byte
	namespace *v1.Namespace
	config    *Config
}

func newHelmApp(o *CreateOptions, projectName, repoURL, targetRevision, repoRoot string) (*helmApp, error) {
	var err error
	app := &helmApp{
		baseApp: baseApp{o},
	}

	if o.AppSpecifier == "" {
		return nil, ErrEmptyAppSpecifier
	}

	if o.AppName == "" {
		o.AppName = "default"
	}

	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	if o.DestNamespace == "" {
		o.DestNamespace = "default"
	}

	if o.InstallationMode != InstallModeFlatten {
		log.G().Warn("helm app only supports flatten installation mode, got: %s", o.InstallationMode)
		o.InstallationMode = InstallModeFlatten
	}

	// parse git url
	_, orgRepo, appPath, _, _, _, _ := util.ParseGitUrl(o.AppSpecifier)
	log.G().WithFields(log.Fields{
		"orgRepo": orgRepo,
		"path":    appPath,
	}).Debug("helm app parsed git url, generating manifests")

	_, appfs, exists := git.GetRepositoryCache().Get(orgRepo, false)
	if !exists {
		return nil, fmt.Errorf("helm app failed to get repository cache")
	}

	manifests, err := dryrun.GenerateHelmManifest(appfs, appPath, "/", "default", o.DestNamespace, o.AppName)
	if err != nil {
		log.G().WithFields(log.Fields{
			"error": err,
		}).Error("helm app generating manifest")
		return nil, err
	}

	app.base = &kusttypes.Kustomization{
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
		Resources: []string{"manifest.yaml"},
	}

	app.overlay = &kusttypes.Kustomization{
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
		Resources: []string{"../../base"},
	}

	app.manifests = manifests

	app.config = &Config{
		AppName:           o.AppName,
		UserGivenName:     o.AppName,
		DestNamespace:     o.DestNamespace,
		DestServer:        o.DestServer,
		SrcPath:           path.Join(repoRoot, store.Default.AppsDir, o.AppName, store.Default.OverlaysDir, projectName),
		SrcRepoURL:        repoURL,
		SrcTargetRevision: targetRevision,
		Labels:            o.Labels,
		Annotations:       o.Annotations,
	}

	return app, nil
}

// CreateHelmApp creates the necessary files for a helm application
func CreateHelmApp(app *helmApp, repofs fs.FS, appsfs fs.FS, projectName string) error {
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

	if err := appsfs.WriteYamls(baseKustomizationPath, app.base); err != nil {
		return err
	}

	// create Overlay
	overlayPath := appsfs.Join(appPath, "overlays", projectName)
	overlayKustomizationPath := appsfs.Join(overlayPath, "kustomization.yaml")
	log.G().Debugf("overlay path: %s", overlayKustomizationPath)
	if appsfs.ExistsOrDie(overlayKustomizationPath) {
		return ErrAppAlreadyInstalledOnProject
	}

	if err := appsfs.WriteYamls(overlayKustomizationPath, app.overlay); err != nil {
		return err
	}

	// create manifests - only used in flat installation mode
	if app.manifests != nil {
		manifestsPath := appsfs.Join(basePath, "manifest.yaml")
		if _, err := writeFile(appsfs, manifestsPath, "manifests", app.manifests); err != nil {
			return err
		}
	}

	clusterName, err := getClusterName(repofs, app.opts.DestServer)
	if err != nil {
		return err
	}

	if app.namespace != nil {
		if err := createNamespaceManifest(repofs, clusterName, app.namespace); err != nil {
			return err
		}
	}

	configPath := repofs.Join(overlayPath, "config.json")
	if repofs != appsfs {
		configPath = repofs.Join(appPath, projectName, "config.json")
	}

	if viper.GetString("gitops.mode") == "pull_request" {
		if err := appsfs.WriteJson(configPath, app.config); err != nil {
			return fmt.Errorf("appsfs failed to write app config.json: %w", err)
		}
		return nil
	}

	if err := repofs.WriteJson(configPath, app.config); err != nil {
		return fmt.Errorf("repofs failed to write app config.json: %w", err)
	}

	return nil
}

func (h *helmApp) CreateFiles(repofs fs.FS, appsfs fs.FS, projectName string) error {
	return CreateHelmApp(h, repofs, appsfs, projectName)
}

func (h *helmApp) Manifests() map[string][]byte {
	return map[string][]byte{
		"default": h.manifests,
	}
}
