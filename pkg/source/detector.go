package source

import (
	"fmt"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

// this file is an abstraction layer for the git repository layout
// supports the following layouts:
// - helm
// - kustomize
// - helm with multiple environments
// - kustomize with multiple environments
// - some existing gitops patterns

func InferApplicationSource(repofs fs.FS, req types.ApplicationSourceRequest) (string, []string, error) {
	appSourceType, err := detectApplicationType(repofs, req)
	if err != nil {
		return "", nil, err
	}

	switch appSourceType {
	case AppTypeHelm, AppTypeKustomize:
		return appSourceType, []string{"default"}, nil

	case AppTypeHelmMultiEnv, AppTypeKustomizeMultiEnv:
		detector := NewEnvironmentDetector(appSourceType, repofs, req.Path)
		environments, err := detector.DetectEnvironments()
		if err != nil {
			return appSourceType, nil, err
		}
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
			"app_source_type": appSourceType,
			"environments":    environments,
		}).Debug("detected application environments")
		return appSourceType, environments, nil

	default:
		return "", nil, fmt.Errorf("unknown application source type: %s", appSourceType)
	}
}

func detectApplicationType(repofs fs.FS, req types.ApplicationSourceRequest) (string, error) {
	log.G().WithFields(log.Fields{
		"repo":            req.Repo,
		"path":            req.Path,
		"target_revision": req.TargetRevision,
	}).Debug("detecting application type")

	files, err := repofs.ReadDir(req.Path)
	if err != nil {
		return "", err
	}
	log.G().Debugf("repofs: %+v", files)

	// 1. application specifier
	// indicates that the user has specified the helm manifest path
	// this is convention for helm applications with multiple environments
	// user tells us which helm manifest path to use
	if req.ApplicationSpecifier.HelmManifestPath != "" {
		log.G().Debug("detected helm application from application specifier")
		return AppTypeHelmMultiEnv, nil
	}

	return inferAppTypeFromPath(repofs, req.Path), nil
}

func InferAppType(repofs fs.FS) string {
	return inferAppTypeFromPath(repofs, repofs.Root())
}

// using heuristic from https://argoproj.github.io/argo-cd/user-guide/tool_detection/#tool-detection
func inferAppTypeFromPath(repofs fs.FS, path string) string {
	if repofs.ExistsOrDie("app.yaml") &&
		repofs.ExistsOrDie("components/params.libsonnet") {
		return AppTypeKsonnet
	}
	// 2. check root path with standard structure
	// 2.1 manifests directory with environments directory
	if repofs.ExistsOrDie(repofs.Join(path, "manifests")) &&
		repofs.ExistsOrDie(repofs.Join(path, "environments")) {
		return AppTypeHelmMultiEnv
	}

	// 2.2 base and overlays directory
	if repofs.ExistsOrDie(repofs.Join(path, "base")) &&
		repofs.ExistsOrDie(repofs.Join(path, "overlays")) {
		return AppTypeKustomizeMultiEnv
	}

	// 3 for simple layout of kustomize or helm
	// 3.1 kustomization.yaml or kustomization.yml with / directory
	if repofs.ExistsOrDie(repofs.Join(path, "kustomization.yaml")) ||
		repofs.ExistsOrDie(repofs.Join(path, "kustomization.yml")) ||
		repofs.ExistsOrDie("Kustomization") {
		return AppTypeKustomize
	}

	// 3.2 Chart.yaml with / directory
	if repofs.ExistsOrDie(repofs.Join(path, "Chart.yaml")) {
		return AppTypeHelm
	}

	log.G().Warnf("could not infer application type from path '%s', use default dir type", path)
	return AppTypeDirectory
}

type EnvironmentDetector interface {
	DetectEnvironments() ([]string, error)
}

// NewEnvironmentDetector creates a new environment detector based on the application source type
func NewEnvironmentDetector(appSourceType string, repofs fs.FS, path string) EnvironmentDetector {
	switch appSourceType {
	case AppTypeHelmMultiEnv:
		return &HelmMultiEnvDetector{repofs: repofs, path: path}
	case AppTypeKustomizeMultiEnv:
		return &KustomizeMultiEnvDetector{repofs: repofs, path: path}
	default:
		return nil
	}
}

type HelmMultiEnvDetector struct {
	repofs fs.FS
	path   string
}

func (d *HelmMultiEnvDetector) DetectEnvironments() ([]string, error) {
	return detectHelmMultiEnv(d.repofs, d.path)
}

func detectHelmMultiEnv(repofs fs.FS, path string) ([]string, error) {
	var envDir string
	if path == "" || path == "/" {
		envDir = "environments"
	} else {
		envDir = repofs.Join(path, "environments")
	}

	if !repofs.ExistsOrDie(envDir) {
		log.G().WithField("path", envDir).Debug("No environments directory found, using default")
		return []string{"default"}, nil
	}

	entries, err := repofs.ReadDir(envDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var environments []string
	for _, entry := range entries {
		if entry.IsDir() {
			envPath := repofs.Join(envDir, entry.Name())
			if repofs.ExistsOrDie(repofs.Join(envPath, "values.yaml")) {
				environments = append(environments, entry.Name())
				log.G().WithField("env", entry.Name()).Debug("Found Helm environment")
			}
		}
	}

	if len(environments) == 0 {
		log.G().Debug("No environments found, using default")
		return []string{"default"}, nil
	}

	return environments, nil
}

type KustomizeMultiEnvDetector struct {
	repofs fs.FS
	path   string
}

func (d *KustomizeMultiEnvDetector) DetectEnvironments() ([]string, error) {
	return detectKustomizeMultiEnv(d.repofs, d.path)
}

func detectKustomizeMultiEnv(repofs fs.FS, path string) ([]string, error) {
	var overlaysDir string
	if path == "" || path == "/" {
		overlaysDir = "overlays"
	} else {
		overlaysDir = repofs.Join(path, "overlays")
	}

	if !repofs.ExistsOrDie(overlaysDir) {
		log.G().WithField("path", overlaysDir).Debug("No overlays directory found, using default")
		return []string{"default"}, nil
	}

	entries, err := repofs.ReadDir(overlaysDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read overlays directory: %w", err)
	}

	var environments []string
	for _, entry := range entries {
		if entry.IsDir() {
			envPath := repofs.Join(overlaysDir, entry.Name())
			if repofs.ExistsOrDie(repofs.Join(envPath, "kustomization.yaml")) ||
				repofs.ExistsOrDie(repofs.Join(envPath, "kustomization.yml")) {
				environments = append(environments, entry.Name())
				log.G().WithField("env", entry.Name()).Debug("Found Kustomize environment")
			}
		}
	}

	if len(environments) == 0 {
		log.G().Debug("No environments found, using default")
		return []string{"default"}, nil
	}

	return environments, nil
}
