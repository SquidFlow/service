package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/yannh/kubeconform/pkg/validator"

	"github.com/squidflow/service/pkg/application/dryrun"
	"github.com/squidflow/service/pkg/application/reposource"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

// ValidateApplicationSourceHandler handles the request for validating application source
func ValidateApplicationSourceHandler(c *gin.Context) {
	var req types.ApplicationSourceRequest
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
	appType, environments, err := reposource.ValidateApplicationStructure(repofs, req)
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
	suiteableEnv := []types.AppSourceWithEnvironment{}

	for _, env := range environments {
		log.G().WithFields(log.Fields{
			"type": appType,
			"env":  env,
		}).Debug("Validating environment")

		envResult := types.AppSourceWithEnvironment{
			Environments: env,
			Valid:        true,
		}

		// generate manifest
		var manifests []byte
		switch appType {
		case reposource.SourceHelm:
		case reposource.SourceHelmMultiEnv:
			manifests, err = dryrun.GenerateHelmManifest(repofs, req, env, "application1", "default")
			if err != nil {
				log.G().WithError(err).Error("Failed to generate helm manifest")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

		case reposource.SourceKustomize:
		case reposource.SourceKustomizeMultiEnv:
			manifests, err = dryrun.GenerateKustomizeManifest(repofs, req, env, "application1", "default")
			if err != nil {
				log.G().WithError(err).Error("Failed to generate kustomize manifest")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}
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

	c.JSON(200, types.ValidateAppSourceResponse{
		Success:      true,
		Message:      fmt.Sprintf("Valid %s application source", appType),
		Type:         string(appType),
		SuiteableEnv: suiteableEnv,
	})
}
