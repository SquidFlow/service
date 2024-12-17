package application

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
	"sigs.k8s.io/kustomize/api/krusty"
	kusttypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
)

// GenerateManifests writes the in-memory kustomization to disk, fixes relative resources and
// runs kustomize build, then returns the generated manifests.
//
// If there is a namespace on 'k' a namespace.yaml file with the namespace object will be
// written next to the persisted kustomization.yaml.
//
// To include the namespace in the generated
// manifests just add 'namespace.yaml' to the resources of the kustomization
func GenerateManifests(k *kusttypes.Kustomization) ([]byte, error) {
	return generateManifests(k)
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

	switch o.InstallationMode {
	case InstallationModeFlat, InstallationModeNormal:
	case "":
		o.InstallationMode = InstallationModeNormal
	default:
		return nil, fmt.Errorf("unknown installation mode: %s", o.InstallationMode)
	}

	app.base = &kusttypes.Kustomization{
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
		Resources: []string{o.AppSpecifier},
	}

	if o.InstallationMode == InstallationModeFlat {
		log.G().Info("building manifests...")
		app.manifests, err = generateManifests(app.base)
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

// fixResourcesPaths adjusts all relative paths in the kustomization file to the specified
// newKustDir.
func fixResourcesPaths(k *kusttypes.Kustomization, newKustDir string) error {
	for i, path := range k.Resources {
		// if path is a remote resource ignore it
		if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
			continue
		}

		absRes, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		k.Resources[i], err = filepath.Rel(newKustDir, absRes)
		log.G().WithFields(log.Fields{
			"from": absRes,
			"to":   k.Resources[i],
		}).Debug("adjusting kustomization paths to local filesystem")
		if err != nil {
			return err
		}
	}

	return nil
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

var generateManifests = func(k *kusttypes.Kustomization) ([]byte, error) {
	td, err := os.MkdirTemp(".", "supervisor-tmp")
	if err != nil {
		return nil, fmt.Errorf("failed creating temp dir: %w", err)
	}
	defer os.RemoveAll(td)

	absTd, err := filepath.Abs(td)
	if err != nil {
		return nil, fmt.Errorf("failed getting abs path for \"%s\": %w", td, err)
	}

	if err = fixResourcesPaths(k, absTd); err != nil {
		return nil, fmt.Errorf("failed fixing resources paths: %w", err)
	}

	kyaml, err := yaml.Marshal(k)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling yaml: %w", err)
	}

	kustomizationPath := filepath.Join(td, "kustomization.yaml")
	if err = os.WriteFile(kustomizationPath, kyaml, 0400); err != nil {
		return nil, fmt.Errorf("failed writing file to \"%s\": %w", kustomizationPath, err)
	}

	log.G().WithFields(log.Fields{
		"bootstrapKustPath": kustomizationPath,
		"resourcePath":      k.Resources[0],
	}).Debugf("running bootstrap kustomization: %s\n", string(kyaml))

	opts := krusty.MakeDefaultOptions()
	kust := krusty.MakeKustomizer(opts)
	fs := filesys.MakeFsOnDisk()
	res, err := kust.Run(fs, td)
	if err != nil {
		return nil, fmt.Errorf("failed running kustomization: %w", err)
	}

	return res.AsYaml()
}
