package application

import (
	"fmt"
	"reflect"

	kusttypes "sigs.k8s.io/kustomize/api/types"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/util"
)

type (
	dirApp struct {
		baseApp
		dirConfig *dirConfig
	}

	dirConfig struct {
		Config
		Exclude string `json:"exclude"`
		Include string `json:"include"`
	}
)

/* dirApp Application impl */
func newDirApp(opts *CreateOptions) *dirApp {
	app := &dirApp{
		baseApp: baseApp{opts},
	}

	host, orgRepo, path, gitRef, _, suffix, _ := util.ParseGitUrl(opts.AppSpecifier)
	url := host + orgRepo + suffix
	if path == "" {
		path = "."
	}

	app.dirConfig = &dirConfig{
		Config: Config{
			AppName:           opts.AppName,
			UserGivenName:     opts.AppName,
			DestNamespace:     opts.DestNamespace,
			DestServer:        opts.DestServer,
			SrcRepoURL:        url,
			SrcPath:           path,
			SrcTargetRevision: gitRef,
			Labels:            opts.Labels,
			Annotations:       opts.Annotations,
		},
		Exclude: opts.Exclude,
		Include: opts.Include,
	}

	return app
}

func (app *dirApp) CreateFiles(repofs fs.FS, appsfs fs.FS, projectName string) error {
	appPath := repofs.Join(store.Default.AppsDir, app.opts.AppName, projectName)
	if repofs.ExistsOrDie(appPath) {
		return ErrAppAlreadyInstalledOnProject
	}

	configPath := repofs.Join(appPath, "config_dir.json")
	if err := repofs.WriteJson(configPath, app.dirConfig); err != nil {
		return fmt.Errorf("failed to write app config_dir.json: %w", err)
	}

	clusterName, err := getClusterName(repofs, app.opts.DestServer)
	if err != nil {
		return err
	}

	if app.opts.DestNamespace != "" && app.opts.DestNamespace != "default" {
		if err = createNamespaceManifest(repofs, clusterName, kube.GenerateNamespace(app.opts.DestNamespace, nil)); err != nil {
			return err
		}
	}

	return nil
}

func (app *dirApp) Manifests() map[string][]byte {
	return nil
}

func writeFile(repofs fs.FS, path, name string, data []byte) (bool, error) {
	absPath := repofs.Join(repofs.Root(), path)
	exists, err := repofs.CheckExistsOrWrite(path, data)
	if err != nil {
		return false, fmt.Errorf("failed to create '%s' file at '%s': %w", name, absPath, err)
	} else if exists {
		log.G().Infof("'%s' file exists in '%s'", name, absPath)
		return true, nil
	}

	log.G().Infof("created '%s' file at '%s'", name, absPath)
	return false, nil
}

func checkBaseCollision(repofs fs.FS, orgBasePath string, newBase *kusttypes.Kustomization) (bool, error) {
	orgBase := &kusttypes.Kustomization{}
	if err := repofs.ReadYamls(orgBasePath, orgBase); err != nil {
		return false, err
	}

	return !reflect.DeepEqual(orgBase, newBase), nil
}
