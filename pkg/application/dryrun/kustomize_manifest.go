package dryrun

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

// GenerateKustomizeManifest generates Kustomize manifests for a specific environment
func GenerateKustomizeManifest(repofs fs.FS, req types.ApplicationSourceRequest, env string, applicationName string, applicationNamespace string) ([]byte, error) {
	log.G().WithFields(log.Fields{
		"path":      req.Path,
		"env":       env,
		"name":      applicationName,
		"namespace": applicationNamespace,
	}).Debug("preparing kustomize build")

	// Create an in-memory filesystem for kustomize
	memFS := filesys.MakeFsInMemory()

	var buildPath string
	switch env {
	// case 1: simple Kustomize, directly use the specified path
	case "default":
		buildPath = req.Path

		// check if kustomization.yaml exists
		if !repofs.ExistsOrDie(repofs.Join(buildPath, "kustomization.yaml")) {
			log.G().WithFields(log.Fields{
				"path": buildPath,
			}).Error("kustomization.yaml not found")
			return nil, fmt.Errorf("kustomization.yaml not found in %s", buildPath)
		}

		// copy the whole directory to memory filesystem
		err := copyToMemFS(repofs, "/", "/", memFS)
		if err != nil {
			log.G().WithError(err).Error("failed to copy files")
			return nil, fmt.Errorf("failed to copy files: %w", err)
		}

	// case 2: multi-environment Kustomize, use overlays structure
	default:
		overlayPath := repofs.Join(req.Path, "overlays", env)
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

		// copy the whole application directory (including base and overlays) to memory filesystem
		err := copyToMemFS(repofs, req.Path, "/", memFS)
		if err != nil {
			log.G().WithError(err).Error("failed to copy files")
			return nil, fmt.Errorf("failed to copy files: %w", err)
		}
		buildPath = repofs.Join("/overlays", env)
	}

	// List files for debugging
	entries, err := memFS.ReadDir(buildPath)
	if err == nil {
		fileNames := make([]string, 0)
		for _, entry := range entries {
			fileNames = append(fileNames, entry)
		}
		log.G().WithFields(log.Fields{
			"path":  buildPath,
			"files": fileNames,
		}).Debug("files in memory filesystem")
	}

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
