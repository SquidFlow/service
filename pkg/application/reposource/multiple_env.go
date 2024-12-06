package reposource

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/squidflow/service/pkg/fs"
)

type EnvironmentDetector interface {
	DetectEnvironments() ([]string, error)
}

// NewEnvironmentDetector creates a new environment detector based on the application source type
func NewEnvironmentDetector(appSourceType AppSourceType, repofs fs.FS, path string) EnvironmentDetector {
	switch appSourceType {
	case SourceHelmMultiEnv:
		return &HelmMultiEnvDetector{repofs: repofs, path: path}
	case SourceKustomizeMultiEnv:
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
	var envDir string
	if d.path == "" || d.path == "/" {
		envDir = "environments"
	} else {
		envDir = d.repofs.Join(d.path, "environments")
	}

	if !d.repofs.ExistsOrDie(envDir) {
		log.WithField("path", envDir).Debug("No environments directory found, using default")
		return []string{"default"}, nil
	}

	entries, err := d.repofs.ReadDir(envDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var environments []string
	for _, entry := range entries {
		if entry.IsDir() {
			envPath := d.repofs.Join(envDir, entry.Name())
			if d.repofs.ExistsOrDie(d.repofs.Join(envPath, "values.yaml")) {
				environments = append(environments, entry.Name())
				log.WithField("env", entry.Name()).Debug("Found Helm environment")
			}
		}
	}

	if len(environments) == 0 {
		log.Debug("No environments found, using default")
		return []string{"default"}, nil
	}

	return environments, nil
}

type KustomizeMultiEnvDetector struct {
	repofs fs.FS
	path   string
}

func (d *KustomizeMultiEnvDetector) DetectEnvironments() ([]string, error) {
	var overlaysDir string
	if d.path == "" || d.path == "/" {
		overlaysDir = "overlays"
	} else {
		overlaysDir = d.repofs.Join(d.path, "overlays")
	}

	if !d.repofs.ExistsOrDie(overlaysDir) {
		log.WithField("path", overlaysDir).Debug("No overlays directory found, using default")
		return []string{"default"}, nil
	}

	entries, err := d.repofs.ReadDir(overlaysDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read overlays directory: %w", err)
	}

	var environments []string
	for _, entry := range entries {
		if entry.IsDir() {
			envPath := d.repofs.Join(overlaysDir, entry.Name())
			if d.repofs.ExistsOrDie(d.repofs.Join(envPath, "kustomization.yaml")) ||
				d.repofs.ExistsOrDie(d.repofs.Join(envPath, "kustomization.yml")) {
				environments = append(environments, entry.Name())
				log.WithField("env", entry.Name()).Debug("Found Kustomize environment")
			}
		}
	}

	if len(environments) == 0 {
		log.Debug("No environments found, using default")
		return []string{"default"}, nil
	}

	return environments, nil
}
