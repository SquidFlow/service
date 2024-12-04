package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/yannh/kubeconform/pkg/validator"

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
	// check first if it is multi-environment structure

	// check Helm multi-environment structure (environments/*)
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

	// check Kustomize multi-environment structure (overlays/*)
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

	// if not multi-environment structure, check standard structure

	// check standard Helm structure
	if repofs.ExistsOrDie(repofs.Join(path, "Chart.yaml")) {
		return validateHelmStructure(repofs, path)
	}

	// check standard Kustomize structure
	if repofs.ExistsOrDie(repofs.Join(path, "kustomization.yaml")) {
		return validateKustomizeStructure(repofs, path)
	}

	// if it is root path, try to find base directory
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
	// check required files
	if !repofs.ExistsOrDie(repofs.Join(path, "Chart.yaml")) {
		return "", nil, fmt.Errorf("missing Chart.yaml in %s", path)
	}

	if !repofs.ExistsOrDie(repofs.Join(path, "templates")) {
		return "", nil, fmt.Errorf("missing templates directory in %s", path)
	}

	// for standard structure, return default as default environment
	return "helm", []string{"default"}, nil
}

// validateKustomizeStructure validates the Kustomize structure
func validateKustomizeStructure(repofs fs.FS, path string) (string, []string, error) {
	// check basic kustomization.yaml
	if !repofs.ExistsOrDie(repofs.Join(path, "kustomization.yaml")) {
		return "", nil, fmt.Errorf("missing kustomization.yaml in %s", path)
	}

	// for standard structure, return default as default environment
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
