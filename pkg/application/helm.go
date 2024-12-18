package application

import (
	"fmt"
	"path"

	v1 "k8s.io/api/core/v1"

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
	namespace *v1.Namespace
	name      string
	manifests []byte

	// config is the config for the helm app
	config *Config
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

	// parse git url
	_, orgRepo, appPath, _, _, _, _ := util.ParseGitUrl(o.AppSpecifier)
	log.G().WithFields(log.Fields{
		"orgRepo": orgRepo,
		"path":    appPath,
	}).Debug("parsed git url, generating helm manifests")

	_, appfs, exists := git.GetRepositoryCache().Get(orgRepo, false)
	if !exists {
		return nil, fmt.Errorf("failed to get repository cache")
	}

	if o.InstallationMode != InstallModeFlatten {
		return nil, fmt.Errorf("helm app does not support installation mode %s", o.InstallationMode)
	}

	log.G().WithFields(log.Fields{
		"path":         appPath,
		"manifestPath": "/",
		"env":          "default",
		"namespace":    o.DestNamespace,
		"name":         o.AppName,
	}).Debug("helm app generating manifest")
	app.manifests, err = dryrun.GenerateHelmManifest(appfs, appPath, "/", "default", o.DestNamespace, o.AppName)
	if err != nil {
		log.G().WithFields(log.Fields{
			"error": err,
		}).Error("helm app generating manifest")
		return nil, err
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
	}).Debug("helm app creating app config")

	return app, nil
}

func (h *helmApp) CreateFiles(repofs fs.FS, appsfs fs.FS, projectName string) error {
	return nil
}

func (h *helmApp) Manifests() map[string][]byte {
	return map[string][]byte{
		"default": h.manifests,
	}
}
