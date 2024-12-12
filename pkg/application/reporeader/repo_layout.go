// this file is a abstraction layer for the git repository layout
// supports the following layouts:
// - helm
// - kustomize
// - helm with multiple environments
// - kustomize with multiple environments
// - some existing gitops patterns
package reporeader

import (
	"fmt"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

type AppSourceType string

const (
	SourceHelm              AppSourceType = "helm"
	SourceKustomize         AppSourceType = "kustomize"
	SourceHelmMultiEnv      AppSourceType = "helm-multiple-env"
	SourceKustomizeMultiEnv AppSourceType = "kustomize-multiple-env"
)

type AppSourceOption struct {
	Repo                 string
	TargetRevision       string
	Path                 string
	Submodules           bool
	ApplicationSpecifier *AppSourceSpecifier
}

// ApplicationSpecifier contains application-specific configuration
type AppSourceSpecifier struct {
	HelmManifestPath string
}

func ValidateApplicationStructure(repofs fs.FS, req types.ApplicationSourceRequest) (AppSourceType, []string, error) {
	appSourceType, err := detectApplicationType(repofs, req)
	if err != nil {
		return "", nil, err
	}

	switch appSourceType {
	case SourceHelm, SourceKustomize:
		return appSourceType, []string{"default"}, nil
	case SourceHelmMultiEnv, SourceKustomizeMultiEnv:
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

func detectApplicationType(repofs fs.FS, req types.ApplicationSourceRequest) (AppSourceType, error) {
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
	if req.ApplicationSpecifier != nil && req.ApplicationSpecifier.HelmManifestPath != "" {
		log.G().Debug("detected helm application from application specifier")
		return SourceHelmMultiEnv, nil
	}

	// 2. check root path with standard structure
	// 2.1 manifests directory with environments directory
	if repofs.ExistsOrDie(repofs.Join(req.Path, "manifests")) &&
		repofs.ExistsOrDie(repofs.Join(req.Path, "environments")) {
		return SourceHelmMultiEnv, nil
	}

	// 2.2 base and overlays directory
	if repofs.ExistsOrDie(repofs.Join(req.Path, "base")) &&
		repofs.ExistsOrDie(repofs.Join(req.Path, "overlays")) {
		return SourceKustomizeMultiEnv, nil
	}

	// 3 for simple layout of kustomize or helm
	// 3.1 kustomization.yaml or kustomization.yml with / directory
	if repofs.ExistsOrDie(repofs.Join(req.Path, "kustomization.yaml")) ||
		repofs.ExistsOrDie(repofs.Join(req.Path, "kustomization.yml")) {
		return SourceKustomize, nil
	}

	// 3.2 Chart.yaml with / directory
	if repofs.ExistsOrDie(repofs.Join(req.Path, "Chart.yaml")) {
		return SourceHelm, nil
	}

	log.G().WithField("path", req.Path).Error("Failed to detect application type")
	return "", fmt.Errorf("could not detect application type at path: %s", req.Path)
}
