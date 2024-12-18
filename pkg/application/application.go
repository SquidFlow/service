package application

import (
	"fmt"

	"github.com/ghodss/yaml"
	billyUtils "github.com/go-git/go-billy/v5/util"
	v1 "k8s.io/api/core/v1"

	"github.com/squidflow/service/pkg/fs"
	reporeader "github.com/squidflow/service/pkg/source"
	"github.com/squidflow/service/pkg/store"
)

//go:generate mockgen -destination=./mocks/application.go -package=mocks -source=./application.go Application

type Application interface {
	Name() string
	CreateFiles(repofs fs.FS, appsfs fs.FS, projectName string) error
	Manifests() map[string][]byte
}

type InstallModeType string

const (
	InstallModeFlatten InstallModeType = "flatten"
	InstallModeNormal  InstallModeType = "normal"
)

func (i InstallModeType) isValid() bool {
	return i == InstallModeFlatten || i == InstallModeNormal
}

type (
	Config struct {
		AppName           string            `json:"appName"`
		UserGivenName     string            `json:"userGivenName"`
		DestNamespace     string            `json:"destNamespace"`
		DestServer        string            `json:"destServer"`
		SrcPath           string            `json:"srcPath"`
		SrcRepoURL        string            `json:"srcRepoURL"`
		SrcTargetRevision string            `json:"srcTargetRevision"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
	}

	ClusterResConfig struct {
		Name   string `json:"name"`
		Server string `json:"server"`
	}

	// CreateOptions is the options for creating an application
	// it's a superset of all the options for all the application types
	CreateOptions struct {
		AppName          string
		AppType          string
		AppSpecifier     string
		DestNamespace    string
		DestServer       string
		InstallationMode InstallModeType
		Labels           map[string]string
		Annotations      map[string]string
		Exclude          string
		Include          string
		AppSource        reporeader.AppSource
	}

	baseApp struct {
		opts *CreateOptions
	}
)

/* CreateOptions impl */
// Parse tries to parse `CreateOptions` into an `Application`.
func (o *CreateOptions) Parse(projectName, repoURL, targetRevision, repoRoot string) (Application, error) {
	switch o.AppType {
	case reporeader.AppTypeKustomize:
		return newKustApp(o, projectName, repoURL, targetRevision, repoRoot)
	case reporeader.AppTypeHelm:
		return newHelmApp(o, projectName, repoURL, targetRevision, repoRoot)
	case reporeader.AppTypeHelmMultiEnv:
		return newHelmMultiEnvApp(o, projectName, repoURL, targetRevision, repoRoot)
	case reporeader.AppTypeKustomizeMultiEnv:
		return newKustWithMultiEnvApp(o, projectName, repoURL, targetRevision, repoRoot)
	case reporeader.AppTypeDirectory:
		return newDirApp(o), nil
	default:
		return nil, ErrUnknownAppType
	}
}

/* baseApp Application impl */
func (app *baseApp) Name() string {
	return app.opts.AppName
}

func getClusterName(repofs fs.FS, destServer string) (string, error) {
	// verify that the dest server already exists
	clusterName, err := serverToClusterName(repofs, destServer)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster name for the specified dest-server: %w", err)
	}
	if clusterName == "" {
		return "", fmt.Errorf("cluster '%s' is not configured yet, you need to create a project that uses this cluster first", destServer)
	}
	return clusterName, nil
}

func createNamespaceManifest(repofs fs.FS, clusterName string, namespace *v1.Namespace) error {
	nsYAML, err := yaml.Marshal(namespace)
	if err != nil {
		return fmt.Errorf("failed to marshal app overlay namespace: %w", err)
	}

	nsPath := repofs.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, clusterName, namespace.Name+"-ns.yaml")
	if _, err = writeFile(repofs, nsPath, "application namespace", nsYAML); err != nil {
		return err
	}
	return nil
}

var getAppRepo = func(repofs fs.FS, appName string) (string, error) {
	overlays, err := billyUtils.Glob(repofs, repofs.Join(store.Default.AppsDir, appName, store.Default.OverlaysDir, "**", "config.json"))
	if err != nil {
		return "", err
	}

	if len(overlays) == 0 {
		return "", fmt.Errorf("Application '%s' has no overlays", appName)
	}

	c := &Config{}
	return c.SrcRepoURL, repofs.ReadJson(overlays[0], c)
}

func serverToClusterName(repofs fs.FS, server string) (string, error) {
	confs, err := billyUtils.Glob(repofs, repofs.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, "*.json"))
	if err != nil {
		return "", err
	}

	for _, confFile := range confs {
		conf := &ClusterResConfig{}
		if err = repofs.ReadYamls(confFile, conf); err != nil {
			return "", err
		}
		if conf.Server == server {
			return conf.Name, nil
		}
	}

	return "", nil
}
