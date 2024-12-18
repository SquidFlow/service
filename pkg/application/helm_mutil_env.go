package application

import (
	"path"

	v1 "k8s.io/api/core/v1"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
)

type helmMultiEnvApp struct {
	baseApp
	name      string
	namespace *v1.Namespace
	config    *Config
	err       map[string]error  // key is the env name
	manifests map[string][]byte // key is the env name
}

func newHelmMultiEnvApp(o *CreateOptions, projectName, repoURL, targetRevision, repoRoot string) (*helmMultiEnvApp, error) {
	var err error

	app := &helmMultiEnvApp{
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

	if app.manifests == nil {
		app.manifests = make(map[string][]byte)
	}

	if app.err == nil {
		app.err = make(map[string]error)
	}

	if app.manifests == nil {
		app.manifests = make(map[string][]byte)
	}

	if app.err == nil {
		app.err = make(map[string]error)
	}

	for _, env := range o.AppSource.DetectEnvironments() {
		app.manifests[env], err = o.AppSource.Manifest(env)
		if err != nil {
			log.G().WithFields(log.Fields{
				"error": err,
			}).Error("helm-multiple-env app generating manifest")
			app.err[env] = err
			continue
		}
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
	}).Debug("helm-multi-env app creating app config")

	return app, nil
}

func (h *helmMultiEnvApp) CreateFiles(repofs fs.FS, appsfs fs.FS, projectName string) error {
	return nil
}
