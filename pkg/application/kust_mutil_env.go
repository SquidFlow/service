package application

import (
	"errors"
	"fmt"

	"github.com/squidflow/service/pkg/application/dryrun"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/util"
)

type kustWithMultiEnvApp struct {
	baseApp
	kustomizePath string
	manifests     map[string][]byte
	err           map[string]error
}

// fake implementation of helm multi env app
func newKustWithMultiEnvApp(o *CreateOptions, projectName, repoURL, targetRevision, repoRoot string) (*kustWithMultiEnvApp, error) {
	var err error
	app := &kustWithMultiEnvApp{
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

	if len(o.Environments) == 0 {
		return nil, fmt.Errorf("helm-multiple-env app requires at least one environment: default")
	}

	if app.manifests == nil {
		app.manifests = make(map[string][]byte)
	}

	if app.err == nil {
		app.err = make(map[string]error)
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
		return nil, fmt.Errorf("kustomize-multiple-env app does not support installation mode %s", o.InstallationMode)
	}

	for _, env := range o.Environments {
		log.G().WithFields(log.Fields{
			"path":      appPath,
			"env":       env,
			"namespace": o.DestNamespace,
			"name":      o.AppName,
		}).Debug("kustomize-multiple-env app generating manifest")
		app.manifests[env], err = dryrun.GenerateKustomizeManifest(appfs, appPath, env)
		if err != nil {
			log.G().WithFields(log.Fields{
				"error": err,
			}).Error("kustomize-multiple-env app generating manifest")
			app.err[env] = err
			continue
		}
	}

	return app, nil
}

func (k *kustWithMultiEnvApp) CreateFiles(repofs fs.FS, appsfs fs.FS, projectName string) error {
	return errors.New("not implemented")
}

func (k *kustWithMultiEnvApp) Manifests() map[string][]byte {
	return k.manifests
}
