package handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yannh/kubeconform/pkg/validator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	customgogit "github.com/h4-poc/service/pkg/git/custom-gogit"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/util"
)

// DryRunRequest represents the request structure for dry run
type DryRunRequest struct {
	Clusters   []string               `json:"clusters" binding:"required"`
	Namespace  string                 `json:"namespace" binding:"required"`
	Template   map[string]interface{} `json:"template" binding:"required"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// DryRunResponse represents the response structure for dry run
type DryRunResponse struct {
	Yamls []ClusterYAML `json:"yamls"`
}

// ClusterYAML represents the generated YAML for each cluster
type ClusterYAML struct {
	Cluster string `json:"cluster"`
	Content string `json:"content"`
}

// ValidateTemplateRequest represents the request structure for template validation
type ValidateTemplateRequest struct {
	TemplateSource string                 `json:"templateSource" binding:"required"`
	TargetRevision string                 `json:"targetRevision" binding:"required"`
	Path           string                 `json:"path,omitempty"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
}

// ValidateTemplateResponse represents the response structure for template validation
type ValidateTemplateResponse struct {
	Results []ValidationResult `json:"results"`
}

func ValidateTemplate(c *gin.Context) {
	var req ValidateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	if err := customgogit.CloneSubModule(req.TemplateSource, req.TargetRevision); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	VResult := []ValidationResult{}
	strList := strings.Split(req.Path, "/")
	app := strList[len(strList)-1]
	entries, err := os.ReadDir(fmt.Sprintf("/tmp/platform/overlays/app/%s", app))
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	envMap := map[string]bool{
		"sit":  true,
		"sit1": true,
		"sit2": true,
		"uat":  true,
		"uat1": true,
		"uat2": true,
	}
	env := []string{}
	for _, e := range entries {
		if e.Type().IsDir() {
			if envMap[e.Name()] {
				env = append(env, e.Name())
			}
		}

	}
	if util.CheckIsHelmChart(fmt.Sprintf("/tmp/platform/manifest/%s/Chart.yaml", app)) {
		for _, env := range env {
			if err := Helm_Templating(app, env); err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
				return
			}
			if err := KustomizeBuildInOverlay(app, env); err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
				return
			}
			errList, err := KubeManifestValidator(fmt.Sprintf("/tmp/platform/overlays/app/%s/%s/generate-manifest.yaml", app, env))
			if err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
				return
			}
			if len(errList) == 0 {
				VResult = append(VResult, ValidationResult{Environment: env, IsValid: true})
			} else {
				VResult = append(VResult, ValidationResult{Environment: env, IsValid: false, Message: errList})
			}

		}

	} else {
		for _, env := range env {
			if err := KustomizeBuildInManifest(app, env); err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
				return
			}
			if err := KustomizeBuildInOverlay(app, env); err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
				return
			}
			errList, err := KubeManifestValidator(fmt.Sprintf("/tmp/platform/overlays/app/%s/%s/generate-manifest.yaml", app, env))
			if err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
				return
			}
			if len(errList) == 0 {
				VResult = append(VResult, ValidationResult{Environment: env, IsValid: true})
			} else {
				VResult = append(VResult, ValidationResult{Environment: env, IsValid: false, Message: errList})
			}
		}
	}
	c.JSON(200, VResult)

}

// DryRunArgoApplications handles the dry run request for Argo applications
func DryRunArgoApplications(c *gin.Context) {
	var req DryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate clusters exist
	for _, cluster := range req.Clusters {
		// Check if cluster exists
		exists, err := validateClusterExists(cluster)
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to validate cluster %s: %v", cluster, err)})
			return
		}
		if !exists {
			c.JSON(404, gin.H{"error": fmt.Sprintf("Cluster %s not found", cluster)})
			return
		}
	}

	response := DryRunResponse{
		Yamls: make([]ClusterYAML, 0, len(req.Clusters)),
	}

	for _, clusterName := range req.Clusters {
		// Deep copy the template for each cluster
		clusterTemplate := deepCopyMap(req.Template)

		// Apply cluster-specific modifications
		if err := applyClusterSpecifics(clusterTemplate, clusterName, req.Namespace, req.Parameters); err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to apply cluster specifics for %s: %v", clusterName, err)})
			return
		}

		// Validate the modified template
		if err := validateDryRunTemplate(clusterTemplate); err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid template for %s: %v", clusterName, err)})
			return
		}

		// Convert to YAML
		yamlContent, err := generateYAML(clusterTemplate)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to generate YAML for %s: %v", clusterName, err)})
			return
		}

		response.Yamls = append(response.Yamls, ClusterYAML{
			Cluster: clusterName,
			Content: yamlContent,
		})
	}

	c.JSON(200, response)
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
