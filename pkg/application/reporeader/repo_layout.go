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
	log.G().WithFields(log.Fields{
		"repo":            req.Path,
		"path":            req.Path,
		"target_revision": req.TargetRevision,
		"app_source_type": appSourceType,
	}).Debug("Detected application type")

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
		}).Debug("Detected application environments")

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
	}).Debug("Detecting application type")

	// the application specifier is used to specify the helm manifest path
	// this is convention for helm applications with multiple environments
	// user tells us which helm manifest path to use
	if req.ApplicationSpecifier != nil && req.ApplicationSpecifier.HelmManifestPath != "" {
		log.G().Debug("Detected Helm application from ApplicationSpecifier")
		return SourceHelmMultiEnv, nil
	}

	// check root path with standard structure
	if repofs.ExistsOrDie(repofs.Join(req.Path, "kustomization.yaml")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Kustomize application from kustomization.yaml")
		return SourceKustomizeMultiEnv, nil
	}

	// this is convention for helm applications
	if repofs.ExistsOrDie(repofs.Join(req.Path, "Chart.yaml")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Helm application from Chart.yaml")
		return SourceHelm, nil
	}

	// this is convention for helm applications with multiple environments
	if repofs.ExistsOrDie(repofs.Join(req.Path, "manifests")) &&
		repofs.ExistsOrDie(repofs.Join(req.Path, "environments")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Helm application from manifests directory")
		return SourceHelmMultiEnv, nil
	}

	// this is convention for kustomize applications
	if repofs.ExistsOrDie(repofs.Join(req.Path, "base")) &&
		repofs.ExistsOrDie(repofs.Join(req.Path, "overlays")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Kustomize application from directories")
		return SourceKustomizeMultiEnv, nil
	}

	log.G().WithField("path", req.Path).Error("Failed to detect application type")
	return "", fmt.Errorf("could not detect application type at path: %s", req.Path)
}
