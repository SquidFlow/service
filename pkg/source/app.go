package source

import (
	"fmt"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/yannh/kubeconform/pkg/validator"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
)

// AppSource represents an application source interface that all app types must implement
type AppSource interface {
	// GetType returns the type of the application source
	GetType() string

	// Validate validates if the application source is valid
	// return map[env]error
	Validate(repofs fs.FS, path string) map[string]error

	// Manifest generates the kubernetes manifests for the given environment
	Manifest(env string) ([]byte, error)

	// DetectEnvironments detects available environments for this application
	DetectEnvironments() []string
}

// AppSourceFactory creates appropriate AppSource based on the application type
func NewAppSource(repofs fs.FS, path string, manifestPath string) (AppSource, error) {
	log.G().WithFields(log.Fields{
		"path":         path,
		"manifestPath": manifestPath,
	}).Debug("new app source")

	if manifestPath != "" {
		return &HelmMultiEnvAppSource{
			repofs:       repofs,
			path:         path,
			manifestPath: manifestPath,
		}, nil
	}

	appType := inferAppTypeFromPath(repofs, path)
	switch appType {
	case AppTypeHelm:
		return &HelmAppSource{
			repofs: repofs,
			path:   path,
		}, nil
	case AppTypeKustomize:
		return &KustomizeAppSource{
			repofs: repofs,
			path:   path,
		}, nil
	case AppTypeHelmMultiEnv:
		return &HelmMultiEnvAppSource{
			repofs:       repofs,
			path:         path,
			manifestPath: manifestPath,
		}, nil
	case AppTypeKustomizeMultiEnv:
		return &KustomizeMultiEnvAppSource{
			repofs: repofs,
			path:   path,
		}, nil
	}
	// skip ksonnet, directory type support
	return nil, fmt.Errorf("unknown app type: %s", appType)
}

// HelmAppSource creates a simple helm app source
type HelmAppSource struct {
	repofs       fs.FS
	path         string
	manifestPath string
}

func (h *HelmAppSource) GetType() string {
	return AppTypeHelm
}

func (h *HelmAppSource) Validate(repofs fs.FS, path string) map[string]error {
	results := make(map[string]error)
	memFS := memfs.New()
	manifest, err := h.Manifest("default")
	if err != nil {
		results["default"] = err
		return results
	}

	if err := validateManifest(memFS, manifest, "default"); err != nil {
		results["default"] = err
		return results
	}
	return results
}

func (h *HelmAppSource) DetectEnvironments() []string {
	return []string{"default"}
}

func (h *HelmAppSource) Manifest(env string) ([]byte, error) {
	manifest, err := GenerateHelmManifest(h.repofs, h.path, h.manifestPath, env, "default", "test-app")
	if err != nil {
		log.G().WithError(err).WithFields(log.Fields{
			"env":  env,
			"path": h.path,
		}).Error("failed to generate helm manifest")
		return nil, err
	}
	return manifest, nil
}

// KustomizeAppSource creates a simple kustomize app source
type KustomizeAppSource struct {
	repofs fs.FS
	path   string
}

func (k *KustomizeAppSource) GetType() string {
	return AppTypeKustomize
}

func (k *KustomizeAppSource) Validate(repofs fs.FS, path string) map[string]error {
	results := make(map[string]error)
	envs := k.DetectEnvironments()

	memFS := memfs.New()
	for _, env := range envs {
		manifest, err := k.Manifest(env)
		if err != nil {
			results[env] = err
			continue
		}
		if err := validateManifest(memFS, manifest, env); err != nil {
			results[env] = err
			continue
		}
	}

	return results
}

func (k *KustomizeAppSource) DetectEnvironments() []string {
	return []string{"default"}
}

func (k *KustomizeAppSource) Manifest(env string) ([]byte, error) {
	manifest, err := GenerateKustomizeManifest(k.repofs, k.path, env)
	if err != nil {
		log.G().WithError(err).WithFields(log.Fields{
			"env":  "default",
			"path": k.path,
		}).Error("failed to generate kustomize manifest")
		return nil, err
	}

	return manifest, nil
}

// HelmMultiEnvAppSource creates a new helm multi env app source
type HelmMultiEnvAppSource struct {
	repofs       fs.FS
	path         string
	manifestPath string
}

func (h *HelmMultiEnvAppSource) GetType() string {
	return AppTypeHelmMultiEnv
}

func (h *HelmMultiEnvAppSource) Validate(repofs fs.FS, path string) map[string]error {
	results := make(map[string]error)
	envs := h.DetectEnvironments()

	memFS := memfs.New()
	for _, env := range envs {
		manifest, err := h.Manifest(env)
		if err != nil {
			results[env] = err
			continue
		}
		if err := validateManifest(memFS, manifest, env); err != nil {
			results[env] = err
			continue
		}
	}

	return results
}

func (h *HelmMultiEnvAppSource) DetectEnvironments() []string {
	envs, err := detectHelmMultiEnv(h.repofs, h.path)
	if err != nil {
		return []string{"default"}
	}
	return envs
}

func (h *HelmMultiEnvAppSource) Manifest(env string) ([]byte, error) {
	manifest, err := GenerateHelmManifest(h.repofs, h.path, h.manifestPath, env, "default", "test-app")
	if err != nil {
		log.G().WithError(err).WithFields(log.Fields{
			"env":  env,
			"path": h.path,
		}).Error("failed to generate helm manifest")
		return nil, err
	}
	return manifest, nil
}

// KustomizeMultiEnvAppSource creates a new kustomize multi env app source
type KustomizeMultiEnvAppSource struct {
	repofs fs.FS
	path   string
}

func (k *KustomizeMultiEnvAppSource) GetType() string {
	return AppTypeKustomizeMultiEnv
}

func (k *KustomizeMultiEnvAppSource) Validate(repofs fs.FS, path string) map[string]error {
	results := make(map[string]error)
	envs := k.DetectEnvironments()

	memFS := memfs.New()
	for _, env := range envs {
		manifest, err := k.Manifest(env)
		if err != nil {
			results[env] = err
			continue
		}
		if err := validateManifest(memFS, manifest, env); err != nil {
			results[env] = err
			continue
		}
	}

	return results
}

func (k *KustomizeMultiEnvAppSource) DetectEnvironments() []string {
	envs, err := detectKustomizeMultiEnv(k.repofs, k.path)
	if err != nil {
		return []string{"default"}
	}
	return envs
}

func (k *KustomizeMultiEnvAppSource) Manifest(env string) ([]byte, error) {
	manifest, err := GenerateKustomizeManifest(k.repofs, k.path, env)
	if err != nil {
		log.G().WithError(err).WithFields(log.Fields{
			"env":  env,
			"path": k.path,
		}).Error("failed to generate kustomize manifest")
		return nil, err
	}
	return manifest, nil
}

func validateManifest(memFS billy.Filesystem, manifests []byte, env string) error {
	manifestPath := fmt.Sprintf("/manifests/%s.yaml", env)
	if err := memFS.MkdirAll("/manifests", 0755); err != nil {
		return fmt.Errorf("failed to create manifests directory: %w", err)
	}

	f, err := memFS.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %w", err)
	}

	if _, err := f.Write(manifests); err != nil {
		f.Close()
		return fmt.Errorf("failed to write manifest: %w", err)
	}
	f.Close()

	// validate manifest with kubeconform
	v, err := validator.New([]string{
		"default",
		"https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json",
	}, validator.Opts{
		Strict:  true,
		Cache:   "/tmp/kubeconform-cache",
		SkipTLS: false,
		Debug:   false,
	})

	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	f, err = memFS.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to open manifest for validation: %w", err)
	}
	defer f.Close()

	results := v.Validate(manifestPath, f)
	for _, res := range results {
		if res.Status == validator.Invalid || res.Status == validator.Error {
			return res.Err
		}
	}

	return nil
}
