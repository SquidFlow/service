package source

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
)

// GenerateHelmManifest generates Helm manifests for a specific environment
func GenerateHelmManifest(repofs fs.FS, path, manifestPath, env, namespace, name string) ([]byte, error) {
	log.G().WithFields(log.Fields{
		"path":         path,
		"manifestPath": manifestPath,
		"env":          env,
		"namespace":    namespace,
		"name":         name,
	}).Debug("preparing helm template")

	if env == "default" {
		return generateSimpleHelm(repofs, path, manifestPath, name, namespace)
	}
	return generateMultiEnvHelm(repofs, path, manifestPath, env, name, namespace)
}

// generateSimpleHelm handles single environment helm builds
func generateSimpleHelm(repofs fs.FS, path, manifestPath string, name, namespace string) ([]byte, error) {
	// determine chart path
	chartPath := path
	if manifestPath != "" {
		chartPath = repofs.Join(path, manifestPath)
	}

	log.G().WithFields(log.Fields{
		"chartPath": chartPath,
	}).Debug("looking for chart")

	// validate Chart.yaml exists
	if !repofs.ExistsOrDie(repofs.Join(chartPath, "Chart.yaml")) {
		return nil, fmt.Errorf("Chart.yaml not found at path: %s", chartPath)
	}

	// read default values
	valuesPath := repofs.Join(chartPath, "values.yaml")
	log.G().WithFields(log.Fields{
		"valuesPath": valuesPath,
	}).Debug("reading default values")

	values, err := readValuesFile(repofs, valuesPath)
	if err != nil {
		return nil, err
	}

	return renderHelmChart(repofs, chartPath, values, name, namespace)
}

// generateMultiEnvHelm handles multi-environment helm builds
func generateMultiEnvHelm(repofs fs.FS, path, manifestPath, env, name, namespace string) ([]byte, error) {
	// determine chart path
	chartPath := path
	if manifestPath != "" {
		chartPath = repofs.Join(path, manifestPath)
	}

	// validate Chart.yaml exists
	if !repofs.ExistsOrDie(repofs.Join(chartPath, "Chart.yaml")) {
		return nil, fmt.Errorf("Chart.yaml not found at path: %s", chartPath)
	}

	// Try environment specific values first
	envValuesPath := repofs.Join(path, "environments", env, "values.yaml")
	var valuesPath string
	if repofs.ExistsOrDie(envValuesPath) {
		valuesPath = envValuesPath
		log.G().WithFields(log.Fields{
			"valuesPath": valuesPath,
		}).Debug("reading environment values")
	} else {
		valuesPath = repofs.Join(chartPath, "values.yaml")
		log.G().WithFields(log.Fields{
			"valuesPath": valuesPath,
		}).Debug("environment values not found, using default values")
	}

	values, err := readValuesFile(repofs, valuesPath)
	if err != nil {
		return nil, err
	}

	return renderHelmChart(repofs, chartPath, values, name, namespace)
}

// readValuesFile reads and parses the values file
func readValuesFile(repofs fs.FS, valuesPath string) (map[string]interface{}, error) {
	valuesContent, err := repofs.ReadFile(valuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %w", err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal(valuesContent, &values); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %w", err)
	}

	return values, nil
}

// renderHelmChart performs the actual helm template rendering
func renderHelmChart(repofs fs.FS, chartPath string, values map[string]interface{}, name, namespace string) ([]byte, error) {
	// create and prepare temp directory
	tmpDir, err := os.MkdirTemp("", "helm-chart-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := copyChartFiles(repofs, chartPath, tmpDir); err != nil {
		return nil, fmt.Errorf("failed to copy chart files: %w", err)
	}

	// create action configuration
	settings := cli.New()
	actionConfig := new(action.Configuration)

	// init action configuration
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		name,
		"secrets",
		log.G().Debugf,
	); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}

	// create install action and configure dry run
	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.ReleaseName = name
	client.Namespace = namespace
	client.ClientOnly = true
	client.SkipCRDs = true
	client.KubeVersion = &chartutil.KubeVersion{
		Version: "v1.28.0",
		Major:   "1",
		Minor:   "28",
	}

	// load chart
	chart, err := loader.Load(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load helm chart: %w", err)
	}

	// execute template rendering
	rel, err := client.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to render templates: %w", err)
	}

	log.G().WithFields(log.Fields{
		"chartPath": chartPath,
		"namespace": namespace,
	}).Debug("Successfully rendered helm templates")

	return []byte(rel.Manifest), nil
}

// Helper function to copy chart files
func copyChartFiles(repofs fs.FS, srcPath, destPath string) error {
	entries, err := repofs.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcFilePath := repofs.Join(srcPath, entry.Name())
		destFilePath := filepath.Join(destPath, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destFilePath, 0755); err != nil {
				return err
			}
			if err := copyChartFiles(repofs, srcFilePath, destFilePath); err != nil {
				return err
			}
			continue
		}

		content, err := repofs.ReadFile(srcFilePath)
		if err != nil {
			return err
		}

		if err := os.WriteFile(destFilePath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

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
