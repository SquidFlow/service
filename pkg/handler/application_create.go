package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/application/dryrun"
	"github.com/squidflow/service/pkg/application/reposource"
	"github.com/squidflow/service/pkg/application/repotarget"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"
	"github.com/squidflow/service/pkg/util"
)

var (
	prepareRepo = func(ctx context.Context, cloneOpts *git.CloneOptions, projectName string) (git.Repository, fs.FS, error) {
		log.G().WithFields(log.Fields{
			"repo-url":      cloneOpts.URL(),
			"repo-revision": cloneOpts.Revision(),
			"repo-path":     cloneOpts.Path(),
		}).Debugf("starting with options:")

		log.G().Infof("cloning git repository: %s", cloneOpts.URL())
		r, repofs, err := getRepo(ctx, cloneOpts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed cloning the repository: %w", err)
		}

		root := repofs.Root()
		log.G().Infof("using revision: \"%s\", installation path: \"%s\"", cloneOpts.Revision(), root)
		if !repofs.ExistsOrDie(store.Default.BootsrtrapDir) {
			return nil, nil, fmt.Errorf("bootstrap directory not found, please execute `repo bootstrap` command")
		}

		if projectName != "" {
			projExists := repofs.ExistsOrDie(repofs.Join(store.Default.ProjectsDir, projectName+".yaml"))
			if !projExists {
				return nil, nil, fmt.Errorf(util.Doc(fmt.Sprintf("project '%[1]s' not found, please execute `<BIN> project create %[1]s`", projectName)))
			}
		}

		log.G().Debug("repository is ok")

		return r, repofs, nil
	}

	getRepo = func(ctx context.Context, cloneOpts *git.CloneOptions) (git.Repository, fs.FS, error) {
		return cloneOpts.GetRepo(ctx)
	}
)

func CreateApplicationHandler(c *gin.Context) {
	username := c.GetString(middleware.UserNameKey)
	tenant := c.GetString(middleware.TenantKey)
	log.G().WithFields(log.Fields{
		"username": username,
		"tenant":   tenant,
	}).Debug("create argo application")

	var createReq types.ApplicationCreateRequest
	if err := c.BindJSON(&createReq); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if tenant != createReq.ApplicationInstantiation.TenantName {
		c.JSON(400, gin.H{"error": "tenant in request body does not match tenant in authorization header"})
		return
	}

	// Handle dry run
	if createReq.IsDryRun {
		result, err := performDryRun(c.Request.Context(), &createReq)
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Dry run failed: %v", err)})
			return
		}
		c.JSON(200, result)
		return
	}

	// Normal application creation flow
	var gitOpsFs = memfs.New()
	var opt = types.AppCreateOptions{
		CloneOpts: &git.CloneOptions{
			Repo:     viper.GetString("application_repo.remote_url"),
			FS:       fs.Create(gitOpsFs),
			Provider: "github",
			Auth: git.Auth{
				Password: viper.GetString("application_repo.access_token"),
			},
			CloneForWrite: false,
		},
		AppsCloneOpts: &git.CloneOptions{
			CloneForWrite: false,
		},
		AppOpts: &application.CreateOptions{
			AppName: createReq.ApplicationInstantiation.ApplicationName,
			AppType: application.AppTypeKustomize,
			AppSpecifier: application.BuildKustomizeResourceRef(application.ApplicationSourceOption{
				Repo:           createReq.ApplicationSource.Repo,
				Path:           createReq.ApplicationSource.Path,
				TargetRevision: createReq.ApplicationSource.TargetRevision,
			}),
			InstallationMode: application.InstallationModeNormal,
			DestServer:       "https://kubernetes.default.svc",
			Annotations: map[string]string{
				"squidflow.github.io/created-by":  username,
				"squidflow.github.io/tenant":      tenant,
				"squidflow.github.io/description": createReq.ApplicationInstantiation.Description,
				"squidflow.github.io/appcode":     createReq.ApplicationInstantiation.AppCode,
			},
		},
		ProjectName: createReq.ApplicationInstantiation.TenantName,
		KubeFactory: kube.NewFactory(),
	}
	opt.CloneOpts.Parse()
	opt.AppsCloneOpts.Parse()

	// TODO: support multiple clusters
	// for _, cluster := range createReq.DestinationClusters.Clusters {
	// 	opt.createOpts.DestServer = cluster

	// 	if err := RunAppCreate(context.Background(), &opt); err != nil {
	// 		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to create application in cluster %s: %v", cluster, err)})
	// 		return
	// 	}
	// }

	var native repotarget.NativeRepoTarget
	if err := native.RunAppCreate(context.Background(), &opt); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to create application in cluster %s: %v", opt.AppOpts.DestServer, err)})
		return
	}

	c.JSON(201, gin.H{
		"message":     "Applications created successfully",
		"application": createReq,
	})
}

func performDryRun(ctx context.Context, req *types.ApplicationCreateRequest) (*types.ApplicationDryRunResult, error) {
	log.G().WithFields(log.Fields{
		"repo":           req.ApplicationSource.Repo,
		"path":           req.ApplicationSource.Path,
		"targetRevision": req.ApplicationSource.TargetRevision,
		"submodules":     req.ApplicationSource.Submodules,
	}).Info("Starting application dry run")

	// Clone repository to get application source
	cloneOpts := &git.CloneOptions{
		Repo:          req.ApplicationSource.Repo,
		FS:            fs.Create(memfs.New()),
		CloneForWrite: false,
		Submodules:    req.ApplicationSource.Submodules,
	}
	cloneOpts.Parse()

	if req.ApplicationSource.TargetRevision != "" {
		cloneOpts.SetRevision(req.ApplicationSource.TargetRevision)
	}

	_, repofs, err := cloneOpts.GetRepo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Detect application type and validate structure
	appType, environments, err := reposource.ValidateApplicationStructure(repofs, req.ApplicationSource)
	if err != nil {
		return nil, err
	}

	log.G().WithFields(log.Fields{
		"type":         appType,
		"environments": environments,
	}).Debug("Detected application structure")

	// Initialize dry run result
	result := &types.ApplicationDryRunResult{
		Success:      true,
		Total:        len(environments),
		Environments: make([]types.ApplicationDryRunEnv, 0, len(environments)),
	}

	// For each environment, render and validate the templates
	for _, env := range environments {
		log.G().WithFields(log.Fields{
			"environment": env,
			"type":        appType,
		}).Debug("Processing environment")

		envResult := types.ApplicationDryRunEnv{
			Environment: env,
			IsValid:     true,
		}

		var manifests []byte
		switch appType {
		case reposource.SourceHelm:
		case reposource.SourceHelmMultiEnv:
			manifests, err = dryrun.GenerateHelmManifest(
				repofs,
				req.ApplicationSource,
				env,
				req.ApplicationInstantiation.ApplicationName,
				req.ApplicationTarget[0].Namespace,
			)
		case reposource.SourceKustomize:
		case reposource.SourceKustomizeMultiEnv:
			manifests, err = dryrun.GenerateKustomizeManifest(
				repofs,
				req.ApplicationSource,
				env,
				req.ApplicationInstantiation.ApplicationName,
				req.ApplicationTarget[0].Namespace,
			)

		default:
			err = fmt.Errorf("unsupported application type: %s", appType)
		}

		if err != nil {
			envResult.IsValid = false
			envResult.Error = err.Error()
			result.Success = false
			log.G().WithError(err).Error("Failed to generate manifest")
		} else {
			envResult.Manifest = string(manifests)
			log.G().Debug("Successfully generated manifest")
		}

		result.Environments = append(result.Environments, envResult)
	}

	if result.Success {
		result.Message = "Successfully generated manifests for all environments"
	} else {
		result.Message = "Failed to generate manifests for some environments"
	}

	log.G().WithField("success", result.Success).Info("Completed application dry run")
	return result, nil
}
