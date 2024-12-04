package handler

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/yannh/kubeconform/pkg/validator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
)

// ValidateApplicationSourceHandler handles the request for validating application source
func ValidateApplicationSourceHandler(c *gin.Context) {
	var req ValidateAppSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Clone repository
	cloneOpts := &git.CloneOptions{
		Repo:          req.Repo,
		FS:            fs.Create(memfs.New()),
		CloneForWrite: false,
	}
	cloneOpts.Parse()

	if req.TargetVersion != "" {
		cloneOpts.SetRevision(req.TargetVersion)
	}

	_, repofs, err := cloneOpts.GetRepo(context.Background())
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": fmt.Sprintf("Failed to clone repository: %v", err),
		})
		return
	}

	// Validate path exists
	if !repofs.ExistsOrDie(req.Path) {
		c.JSON(400, gin.H{
			"success": false,
			"message": fmt.Sprintf("Path %s does not exist in repository", req.Path),
		})
		return
	}

	// Detect application type and validate structure
	appType, environments, err := validateApplicationStructure(repofs, req.Path)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, ValidateAppSourceResponse{
		Success:      true,
		Message:      fmt.Sprintf("Valid %s application source", appType),
		Type:         appType,
		SuiteableEnv: environments,
	})
}

// validateApplicationStructure validates the application structure
func validateApplicationStructure(repofs fs.FS, path string) (string, []string, error) {
	// First check for multi-environment structure

	// Check for Helm multi-environment structure (environments/*)
	envDir := repofs.Join(path, "environments")
	if repofs.ExistsOrDie(envDir) {
		environments, err := detectHelmEnvironments(repofs, envDir)
		if err != nil {
			return "", nil, err
		}
		if len(environments) > 0 {
			return "helm", environments, nil
		}
	}

	// Check for Kustomize multi-environment structure (overlays/*)
	overlaysDir := repofs.Join(path, "overlays")
	if repofs.ExistsOrDie(overlaysDir) {
		environments, err := detectKustomizeEnvironments(repofs, overlaysDir)
		if err != nil {
			return "", nil, err
		}
		if len(environments) > 0 {
			return "kustomize", environments, nil
		}
	}

	// If not multi-environment, check for standard structure

	// Check for standard Helm structure
	if repofs.ExistsOrDie(repofs.Join(path, "Chart.yaml")) {
		return validateHelmStructure(repofs, path)
	}

	// Check for standard Kustomize structure
	if repofs.ExistsOrDie(repofs.Join(path, "kustomization.yaml")) {
		return validateKustomizeStructure(repofs, path)
	}

	// If root path, try to find in base directory
	if path == "/" || path == "" {
		basePath := repofs.Join(path, "base")
		if repofs.ExistsOrDie(basePath) {
			if repofs.ExistsOrDie(repofs.Join(basePath, "kustomization.yaml")) {
				return validateKustomizeStructure(repofs, basePath)
			}
		}
	}

	return "", nil, fmt.Errorf("no valid application structure found in path %s", path)
}

// validateHelmStructure validates the Helm structure
func validateHelmStructure(repofs fs.FS, path string) (string, []string, error) {
	// Check required files
	if !repofs.ExistsOrDie(repofs.Join(path, "Chart.yaml")) {
		return "", nil, fmt.Errorf("missing Chart.yaml in %s", path)
	}

	if !repofs.ExistsOrDie(repofs.Join(path, "templates")) {
		return "", nil, fmt.Errorf("missing templates directory in %s", path)
	}

	// For standard structure, return "default" as the environment
	return "helm", []string{"default"}, nil
}

// validateKustomizeStructure validates the Kustomize structure
func validateKustomizeStructure(repofs fs.FS, path string) (string, []string, error) {
	// Check for basic kustomization.yaml
	if !repofs.ExistsOrDie(repofs.Join(path, "kustomization.yaml")) {
		return "", nil, fmt.Errorf("missing kustomization.yaml in %s", path)
	}

	// For standard structure, return "default" as the environment
	return "kustomize", []string{"default"}, nil
}

// detectHelmEnvironments detects the environments for Helm
func detectHelmEnvironments(repofs fs.FS, envDir string) ([]string, error) {
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
			}
		}
	}

	return environments, nil
}

// detectKustomizeEnvironments detects the environments for Kustomize
func detectKustomizeEnvironments(repofs fs.FS, overlaysDir string) ([]string, error) {
	entries, err := repofs.ReadDir(overlaysDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read overlays directory: %w", err)
	}

	var environments []string
	for _, entry := range entries {
		if entry.IsDir() {
			envPath := repofs.Join(overlaysDir, entry.Name())
			if repofs.ExistsOrDie(repofs.Join(envPath, "kustomization.yaml")) {
				environments = append(environments, entry.Name())
			}
		}
	}

	return environments, nil
}

// validateClusterExists checks if the given cluster exists
func validateClusterExists(clusterName string) (bool, error) {
	// TODO: Implement actual cluster validation logic
	// This should check against your cluster store/database
	return true, nil
}

// validateDryRunTemplate validates the modified template for dry run
func validateDryRunTemplate(template map[string]interface{}) error {
	// Validate required fields
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	// Check for required top-level fields
	requiredFields := []string{"apiVersion", "kind", "metadata"}
	for _, field := range requiredFields {
		if _, exists := template[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// Validate metadata
	metadata, ok := template["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("metadata must be an object")
	}

	// Check for required metadata fields
	if _, exists := metadata["name"]; !exists {
		return fmt.Errorf("metadata.name is required")
	}

	// Validate spec if it exists
	if spec, exists := template["spec"].(map[string]interface{}); exists {
		if err := validateSpec(spec); err != nil {
			return fmt.Errorf("invalid spec: %w", err)
		}
	}

	return nil
}

// validateSpec validates the spec section of the template
func validateSpec(spec map[string]interface{}) error {
	// Add your spec validation logic here
	// This is just an example - adjust according to your needs
	requiredSpecFields := []string{"source", "destination"}
	for _, field := range requiredSpecFields {
		if _, exists := spec[field]; !exists {
			return fmt.Errorf("required spec field '%s' is missing", field)
		}
	}
	return nil
}

// deepCopyMap creates a deep copy of the template map
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	// Convert to unstructured and back for deep copy
	obj := &unstructured.Unstructured{Object: m}
	return obj.DeepCopy().Object
}

// applyClusterSpecifics modifies the template for specific cluster
func applyClusterSpecifics(template map[string]interface{}, clusterName, namespace string, params map[string]interface{}) error {
	// Set cluster-specific values
	metadata, ok := template["metadata"].(map[string]interface{})
	if !ok {
		metadata = make(map[string]interface{})
		template["metadata"] = metadata
	}

	// Set namespace
	metadata["namespace"] = namespace

	// Apply cluster-specific labels
	labels, ok := metadata["labels"].(map[string]interface{})
	if !ok {
		labels = make(map[string]interface{})
		metadata["labels"] = labels
	}
	labels["cluster"] = clusterName

	// Apply any cluster-specific parameters
	if params != nil {
		if err := applyParameters(template, params); err != nil {
			return fmt.Errorf("failed to apply parameters: %w", err)
		}
	}

	return nil
}

// applyParameters applies the provided parameters to the template
func applyParameters(template map[string]interface{}, params map[string]interface{}) error {
	// Implement recursive parameter substitution
	for key, value := range params {
		if err := applyParameter(template, key, value); err != nil {
			return err
		}
	}
	return nil
}

// applyParameter applies a single parameter to the template
func applyParameter(obj map[string]interface{}, key string, value interface{}) error {
	// Handle nested parameters using dot notation (e.g., "spec.replicas")
	parts := strings.Split(key, ".")
	current := obj

	// Navigate to the parent object
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part].(map[string]interface{})
		if !ok {
			next = make(map[string]interface{})
			current[part] = next
		}
		current = next
	}

	// Set the value
	current[parts[len(parts)-1]] = value
	return nil
}

// generateYAML converts the template to YAML string
func generateYAML(template map[string]interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(template)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template to YAML: %w", err)
	}
	return string(yamlBytes), nil
}

func KubeManifestValidator(generateManifestPath string) ([]string, error) {
	f, err := os.Open(generateManifestPath)
	if err != nil {
		return nil, err
	}
	v, err := validator.New([]string{"default", "https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json"}, validator.Opts{Strict: true, Cache: "/tmp/kubeconform"})
	errList := []string{}
	for _, res := range v.Validate(generateManifestPath, f) {
		if res.Status == validator.Invalid {
			log.G().Info(res.Err.Error())
			errList = append(errList, res.Err.Error())
		}
		if res.Status == validator.Error {
			log.G().Info(res.Err.Error())
			errList = append(errList, res.Err.Error())
		}
	}
	return errList, nil
}
