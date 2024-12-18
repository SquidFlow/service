package application

import (
	"errors"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
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
