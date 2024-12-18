package reader

import (
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

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
