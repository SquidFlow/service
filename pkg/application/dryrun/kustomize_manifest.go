package dryrun

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
)

// GenerateKustomizeManifest generates Kustomize manifests for a specific environment
func GenerateKustomizeManifest(repofs fs.FS, path string, env string) ([]byte, error) {
	log.G().WithFields(log.Fields{
		"repo": repofs,
		"path": path,
		"env":  env,
	}).Debug("preparing kustomize build")

	if env == "default" {
		return generateSimpleKustomize(repofs, path)
	}
	return generateMultiEnvKustomize(repofs, path, env)
}

// generateSimpleKustomize handles single environment kustomize builds
func generateSimpleKustomize(repofs fs.FS, buildPath string) ([]byte, error) {
	// check if kustomization.yaml exists
	kustomizationPath := repofs.Join(buildPath, "kustomization.yaml")
	if !repofs.ExistsOrDie(kustomizationPath) {
		log.G().WithFields(log.Fields{
			"path": kustomizationPath,
		}).Error("kustomization.yaml not found")
		return nil, fmt.Errorf("kustomization.yaml not found in %s", buildPath)
	}

	// Create an in-memory filesystem for kustomize
	memFS := filesys.MakeFsInMemory()

	// copy the whole directory to memory filesystem
	if err := copyToMemFS(repofs, "/", "/", memFS); err != nil {
		log.G().WithError(err).Error("failed to copy files")
		return nil, fmt.Errorf("failed to copy files: %w", err)
	}

	return buildKustomize(memFS, buildPath)
}

// generateMultiEnvKustomize handles multi-environment kustomize builds
func generateMultiEnvKustomize(repofs fs.FS, basePath string, env string) ([]byte, error) {
	overlayPath := repofs.Join(basePath, "overlays", env)
	if !repofs.ExistsOrDie(overlayPath) {
		log.G().WithFields(log.Fields{
			"path": overlayPath,
		}).Error("overlay directory for environment not found")
		return nil, fmt.Errorf("overlay directory for environment %s not found", env)
	}

	// check if kustomization.yaml exists in overlay
	if !repofs.ExistsOrDie(repofs.Join(overlayPath, "kustomization.yaml")) {
		log.G().WithFields(log.Fields{
			"path": overlayPath,
		}).Error("kustomization.yaml not found in overlay")
		return nil, fmt.Errorf("kustomization.yaml not found in overlay %s", env)
	}

	// Create an in-memory filesystem for kustomize
	memFS := filesys.MakeFsInMemory()

	// copy the whole application directory (including base and overlays) to memory filesystem
	if err := copyToMemFS(repofs, basePath, "/", memFS); err != nil {
		log.G().WithError(err).Error("failed to copy files")
		return nil, fmt.Errorf("failed to copy files: %w", err)
	}

	return buildKustomize(memFS, repofs.Join("/overlays", env))
}

// buildKustomize performs the actual kustomize build operation
func buildKustomize(memFS filesys.FileSystem, buildPath string) ([]byte, error) {
	// Create kustomize build options
	opts := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(opts)

	// Build manifests using the in-memory filesystem
	m, err := k.Run(memFS, buildPath)
	if err != nil {
		log.G().WithFields(log.Fields{
			"error": err,
			"path":  buildPath,
		}).Error("failed to build kustomize")
		return nil, fmt.Errorf("failed to build kustomize: %w", err)
	}

	// Get YAML output
	yaml, err := m.AsYaml()
	if err != nil {
		log.G().WithFields(log.Fields{
			"error": err,
			"path":  buildPath,
		}).Error("failed to generate yaml")
		return nil, fmt.Errorf("failed to generate yaml: %w", err)
	}

	log.G().Debug("successfully generated kustomize manifest")
	return yaml, nil
}
