package dryrun

import (
	"fmt"

	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/fs"
)

// GenerateKustomizeManifest generates Kustomize manifests for a specific environment
func GenerateKustomizeManifest(repofs fs.FS, req types.ApplicationSourceRequest, env string, applicationName string, applicationNamespace string) ([]byte, error) {
	log.G().WithFields(log.Fields{
		"path":      req.Path,
		"env":       env,
		"name":      applicationName,
		"namespace": applicationNamespace,
	}).Debug("Preparing kustomize build")

	// Create an in-memory filesystem for kustomize
	memFS := filesys.MakeFsInMemory()

	// configure build path
	var buildPath string
	if env == "default" {
		// case 1: simple Kustomize, directly use the specified path
		buildPath = req.Path

		// check if kustomization.yaml exists
		if !repofs.ExistsOrDie(repofs.Join(buildPath, "kustomization.yaml")) {
			return nil, fmt.Errorf("kustomization.yaml not found in %s", buildPath)
		}

		// copy the whole directory to memory filesystem
		err := copyToMemFS(repofs, buildPath, "/", memFS)
		if err != nil {
			return nil, fmt.Errorf("failed to copy files: %w", err)
		}
		buildPath = "/"
	} else {
		// case 2: multi-environment Kustomize, use overlays structure
		overlayPath := repofs.Join(req.Path, "overlays", env)
		if !repofs.ExistsOrDie(overlayPath) {
			return nil, fmt.Errorf("overlay directory for environment %s not found", env)
		}

		// check if kustomization.yaml exists in overlay
		if !repofs.ExistsOrDie(repofs.Join(overlayPath, "kustomization.yaml")) {
			return nil, fmt.Errorf("kustomization.yaml not found in overlay %s", env)
		}

		// copy the whole application directory (including base and overlays) to memory filesystem
		err := copyToMemFS(repofs, req.Path, "/", memFS)
		if err != nil {
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
		}).Debug("Files in memory filesystem")
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
		}).Error("Failed to build kustomize")
		return nil, fmt.Errorf("failed to build kustomize: %w", err)
	}

	// Get YAML output
	yaml, err := m.AsYaml()
	if err != nil {
		return nil, fmt.Errorf("failed to generate yaml: %w", err)
	}

	log.G().Debug("Successfully generated kustomize manifest")
	return yaml, nil
}
