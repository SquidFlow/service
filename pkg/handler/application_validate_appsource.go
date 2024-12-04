package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/yannh/kubeconform/pkg/validator"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
)

// ValidateApplicationSourceHandler handles the request for validating application source
func ValidateApplicationSourceHandler(c *gin.Context) {
	var req ApplicationSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// set default path
	if req.Path == "" {
		req.Path = "/"
	}

	// set default revision
	if req.TargetRevision == "" {
		req.TargetRevision = "main"
	}

	log.G().WithFields(log.Fields{
		"repo":     req.Repo,
		"path":     req.Path,
		"revision": req.TargetRevision,
	}).Info("Starting application source validation")

	// Clone repository
	cloneOpts := &git.CloneOptions{
		Repo:          req.Repo,
		FS:            fs.Create(memfs.New()),
		CloneForWrite: false,
		Submodules:    req.Submodules,
	}
	cloneOpts.Parse()

	if req.TargetRevision != "" {
		cloneOpts.SetRevision(req.TargetRevision)
	}

	_, repofs, err := cloneOpts.GetRepo(context.Background())
	if err != nil {
		log.G().WithError(err).Error("Failed to clone repository")
		c.JSON(400, gin.H{
			"success": false,
			"message": fmt.Sprintf("Failed to clone repository: %v", err),
		})
		return
	}

	// Detect application type and validate structure
	appType, environments, err := validateApplicationStructure(repofs, req)
	if err != nil {
		log.G().WithError(err).Error("Failed to validate application structure")
		c.JSON(400, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	log.G().WithFields(log.Fields{
		"repo":         req.Repo,
		"path":         req.Path,
		"revision":     req.TargetRevision,
		"type":         appType,
		"environments": environments,
	}).Info("Detected application structure")

	memFS := memfs.New()
	suiteableEnv := []AppSourceWithEnvironment{}

	for _, env := range environments {
		log.G().WithFields(log.Fields{
			"type": appType,
			"env":  env,
		}).Debug("Validating environment")

		envResult := AppSourceWithEnvironment{
			Environments: env,
			Valid:        true,
		}

		// generate manifest
		var manifests []byte
		switch appType {
		case "helm":
			manifests, err = generateHelmManifest(repofs, &req, env, "application1", "default")
		case "kustomize":
			manifests, err = generateKustomizeManifest(repofs, &req, env, "application1", "default")
		}

		if err != nil {
			log.G().WithError(err).WithField("env", env).Error("Failed to generate manifest")
			envResult.Valid = false
			envResult.Error = err.Error()
		} else {
			// write manifest to memory file system
			manifestPath := fmt.Sprintf("/manifests/%s.yaml", env)
			if err := memFS.MkdirAll("/manifests", 0755); err != nil {
				log.G().WithError(err).Error("Failed to create manifests directory")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

			f, err := memFS.Create(manifestPath)
			if err != nil {
				log.G().WithError(err).Error("Failed to create manifest file")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

			if _, err := f.Write(manifests); err != nil {
				f.Close()
				log.G().WithError(err).Error("Failed to write manifest")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
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
				log.G().WithError(err).Error("Failed to create validator")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

			f, err = memFS.Open(manifestPath)
			if err != nil {
				log.G().WithError(err).Error("Failed to open manifest for validation")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

			results := v.Validate(manifestPath, f)
			f.Close()

			for _, res := range results {
				if res.Status == validator.Invalid || res.Status == validator.Error {
					envResult.Valid = false
					envResult.Error = res.Err.Error()
					log.G().WithFields(log.Fields{
						"env":   env,
						"error": res.Err.Error(),
					}).Error("Manifest validation failed")
					break
				}
			}
		}

		suiteableEnv = append(suiteableEnv, envResult)
	}

	c.JSON(200, ValidateAppSourceResponse{
		Success:      true,
		Message:      fmt.Sprintf("Valid %s application source", appType),
		Type:         appType,
		SuiteableEnv: suiteableEnv,
	})
}

func validateApplicationStructure(repofs fs.FS, req ApplicationSourceRequest) (string, []string, error) {
	appType, err := detectApplicationType(repofs, req)
	if err != nil {
		return "", nil, err
	}
	environments := []string{}
	if appType == "helm" {
		environments, err = detectHelmEnvironments(repofs, req.Path)
	} else {
		environments, err = detectKustomizeEnvironments(repofs, req.Path)
	}

	return appType, environments, nil
}

func detectApplicationType(repofs fs.FS, req ApplicationSourceRequest) (string, error) {
	log.G().WithFields(log.Fields{
		"repo":            req.Repo,
		"path":            req.Path,
		"target_revision": req.TargetRevision,
	}).Debug("Detecting application type")

	// the application specifier is used to specify the helm manifest path
	if req.ApplicationSpecifier != nil && req.ApplicationSpecifier.HelmManifestPath != "" {
		log.G().Debug("Detected Helm application from ApplicationSpecifier")
		return "helm", nil
	}

	// check root path with standard structure
	if repofs.ExistsOrDie(repofs.Join(req.Path, "kustomization.yaml")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Kustomize application from kustomization.yaml")
		return "kustomize", nil
	}

	if repofs.ExistsOrDie(repofs.Join(req.Path, "Chart.yaml")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Helm application from Chart.yaml")
		return "helm", nil
	}

	// multiple environments
	// this is convention for helm applications
	if repofs.ExistsOrDie(repofs.Join(req.Path, "manifests")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Helm application from manifests directory")
		return "helm", nil
	}

	// this is convention for kustomize applications
	if repofs.ExistsOrDie(repofs.Join(req.Path, "base")) &&
		repofs.ExistsOrDie(repofs.Join(req.Path, "overlays")) {
		log.G().WithFields(log.Fields{
			"repo":            req.Repo,
			"path":            req.Path,
			"target_revision": req.TargetRevision,
		}).Debug("Detected Kustomize application from directories")
		return "kustomize", nil
	}

	log.G().WithField("path", req.Path).Error("Failed to detect application type")
	return "", fmt.Errorf("could not detect application type at path: %s", req.Path)
}

// detectHelmEnvironments detects the environments for Helm
func detectHelmEnvironments(repofs fs.FS, path string) ([]string, error) {
	var envDir string
	if path == "" || path == "/" {
		envDir = "environments"
	} else {
		envDir = repofs.Join(path, "environments")
	}

	if !repofs.ExistsOrDie(envDir) {
		log.G().WithField("path", envDir).Debug("No environments directory found, using default")
		return []string{"default"}, nil
	}

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
				log.G().WithField("env", entry.Name()).Debug("Found Helm environment")
			}
		}
	}

	if len(environments) == 0 {
		log.G().Debug("No environments found, using default")
		return []string{"default"}, nil
	}

	return environments, nil
}

// detectKustomizeEnvironments detects the environments for Kustomize
func detectKustomizeEnvironments(repofs fs.FS, path string) ([]string, error) {
	var overlaysDir string
	if path == "" || path == "/" {
		overlaysDir = "overlays"
	} else {
		overlaysDir = repofs.Join(path, "overlays")
	}

	if !repofs.ExistsOrDie(overlaysDir) {
		log.G().WithField("path", overlaysDir).Debug("No overlays directory found, using default")
		return []string{"default"}, nil
	}

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
				log.G().WithField("env", entry.Name()).Debug("Found Kustomize environment")
			}
		}
	}

	if len(environments) == 0 {
		log.G().Debug("No environments found, using default")
		return []string{"default"}, nil
	}

	return environments, nil
}
