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

// ValidateAppSource support helm chart, kustomize
func ValidateApplicationSourceHandler(c *gin.Context) {
	var req ValidateAppSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Create clone options
	cloneOpts := &git.CloneOptions{
		Repo:          req.Repo,
		FS:            fs.Create(memfs.New()),
		CloneForWrite: false,
	}
	cloneOpts.Parse()

	if req.TargetVersion != "" {
		cloneOpts.SetRevision(req.TargetVersion)
	}

	// Clone repository into memory fs
	log.G().WithFields(log.Fields{
		"repo":     req.Repo,
		"revision": req.TargetVersion,
		"path":     req.Path,
	}).Info("ValidateAppSource")

	_, repofs, err := cloneOpts.GetRepo(context.Background())
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to clone repository: %v", err)})
		return
	}

	if !repofs.ExistsOrDie(req.Path) {
		err := fmt.Errorf("path %s does not exist in repository %s", req.Path, req.Repo)
		log.G().Error(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Check if path exists
	if !repofs.ExistsOrDie(req.Path) {
		err := fmt.Errorf("path %s does not exist in repository %s", req.Path, req.Repo)
		log.G().Error(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Check application type (Helm vs Kustomize)
	chartPath := repofs.Join(req.Path, "Chart.yaml")
	kustomizationPath := repofs.Join(req.Path, "kustomization.yaml")

	var appType string
	if repofs.ExistsOrDie(chartPath) {
		appType = "helm"
	} else if repofs.ExistsOrDie(kustomizationPath) {
		appType = "kustomize"
	} else {
		c.JSON(400, gin.H{"error": "Neither Chart.yaml nor kustomization.yaml found in specified path"})
		return
	}

	// Validate manifest files
	manifestPath := repofs.Join(req.Path, "templates")
	if appType == "helm" {
		if !repofs.ExistsOrDie(manifestPath) {
			c.JSON(400, gin.H{"error": "Helm chart templates directory not found"})
			return
		}
	}

	// Return validation result
	c.JSON(200, gin.H{
		"isValid": true,
		"type":    appType,
		"message": fmt.Sprintf("Valid %s application source", appType),
	})
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
