package handler

import (
	"context"
	"fmt"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	"github.com/yannh/kubeconform/pkg/validator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/application/dryrun"
	"github.com/squidflow/service/pkg/application/reporeader"
	"github.com/squidflow/service/pkg/application/repowriter"
	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"
)

func ApplicationCreate(c *gin.Context) {
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
			c.JSON(400, gin.H{"error": fmt.Sprintf("dry run failed: %v", err)})
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
				argocd.AnnotationKeyEnvironment: username,
				argocd.AnnotationKeyTenant:      tenant,
				argocd.AnnotationKeyDescription: createReq.ApplicationInstantiation.Description,
				argocd.AnnotationKeyAppCode:     createReq.ApplicationInstantiation.AppCode,
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

	if err := repowriter.Repo().RunAppCreate(context.Background(), &opt); err != nil {
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
	appType, environments, err := reporeader.ValidateApplicationStructure(repofs, req.ApplicationSource)
	if err != nil {
		log.G().WithError(err).Error("failed to validate application structure")
		return nil, err
	}

	log.G().WithFields(log.Fields{
		"type":         appType,
		"environments": environments,
	}).Debug("detected application structure")

	// Initialize dry run result
	result := &types.ApplicationDryRunResult{
		Success:      true,
		Total:        len(environments),
		Environments: make([]types.ApplicationDryRunEnv, 0, len(environments)),
	}

	// For each environment, render and validate the templates
	for _, env := range environments {
		log.G().WithFields(log.Fields{
			"environment":      env,
			"appliction type":  appType,
			"source repo":      req.ApplicationSource.Repo,
			"source path":      req.ApplicationSource.Path,
			"target namespace": req.ApplicationTarget[0].Namespace,
			"target app name":  req.ApplicationInstantiation.ApplicationName,
		}).Debug("processing dry run parameters")

		envResult := types.ApplicationDryRunEnv{
			Environment: env,
			IsValid:     true,
			Manifest:    ``,
			Error:       "",
		}

		var manifests []byte
		switch appType {
		case reporeader.SourceHelm, reporeader.SourceHelmMultiEnv:
			log.G().Debug("generating helm manifest")
			manifests, err = dryrun.GenerateHelmManifest(
				repofs,
				req.ApplicationSource,
				env,
				req.ApplicationInstantiation.ApplicationName,
				req.ApplicationTarget[0].Namespace,
			)
		case reporeader.SourceKustomize, reporeader.SourceKustomizeMultiEnv:
			log.G().Debug("generating kustomize manifest")
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
			result.Success = false
			envResult.IsValid = false
			envResult.Error = err.Error()
			log.G().WithError(err).Error("failed to generate manifest")
		} else {
			envResult.IsValid = true
			envResult.Manifest = string(manifests)
			envResult.Error = ""
			log.G().Debug("successfully generated manifest")
		}

		result.Environments = append(result.Environments, envResult)
	}

	if result.Success {
		result.Message = "successfully generated manifests for all environments"
	} else {
		result.Message = "failed to generate manifests for some environments"
	}

	log.G().WithField("success", result.Success).Info("completed application dry run")
	return result, nil
}

func ApplicationDelete(c *gin.Context) {
	username := c.GetString(middleware.UserNameKey)
	tenant := c.GetString(middleware.TenantKey)
	appName := c.Param("name")

	log.G().WithFields(log.Fields{
		"username": username,
		"tenant":   tenant,
		"appName":  appName,
	}).Debug("delete argo application")

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	applicationName := fmt.Sprintf("%s-%s", tenant, appName)
	_, err = argoClient.Applications(store.Default.ArgoCDNamespace).Get(context.Background(), applicationName, metav1.GetOptions{})
	if err != nil {
		c.JSON(404, gin.H{"error": fmt.Sprintf("Application not found: %v", err)})
		return
	}

	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: true,
	}
	cloneOpts.Parse()

	if err := repowriter.Repo().RunAppDelete(context.Background(), &types.AppDeleteOptions{
		CloneOpts:   cloneOpts,
		ProjectName: tenant,
		AppName:     appName,
		Global:      false,
	}); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete application: %v", err)})
		return
	}

	c.JSON(204, nil)
}

func ApplicationGet(c *gin.Context) {
	tenant := c.GetString(middleware.TenantKey)
	username := c.GetString(middleware.UserNameKey)
	appName := c.Param("name")

	log.G().Infof("tenant: %s, username: %s, appName: %s", tenant, username, appName)

	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: false,
	}
	cloneOpts.Parse()

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	app, err := repowriter.Repo().RunAppGet(context.Background(), &types.AppListOptions{
		CloneOpts:    cloneOpts,
		ProjectName:  tenant,
		ArgoCDClient: argoClient,
	}, appName)

	var argocdappname = fmt.Sprintf("%s-%s", app.ApplicationInstantiation.TenantName, app.ApplicationInstantiation.ApplicationName)
	log.G().WithFields(log.Fields{
		"application namespace": app.ApplicationInstantiation.TenantName,
		"application name":      store.Default.ArgoCDNamespace,
	}).Debug("get application status")

	//TODO: opt with list method
	applicationRuntime, err := argoClient.Applications(store.Default.ArgoCDNamespace).
		Get(context.Background(), argocdappname, metav1.GetOptions{})
	if err != nil {
		log.G().WithError(err).Error("Failed to get application")
	} else {
		app.ApplicationRuntime.Status = getAppStatus(applicationRuntime)
		app.ApplicationRuntime.Health = getAppHealth(applicationRuntime)
		app.ApplicationRuntime.SyncStatus = getAppSyncStatus(applicationRuntime)
		app.ApplicationRuntime.ArgoCDUrl = fmt.Sprintf("https://argocd.squidflow.io/applications/%s", argocdappname)
		app.ApplicationRuntime.CreatedAt = applicationRuntime.CreationTimestamp.Time
		app.ApplicationRuntime.CreatedBy = applicationRuntime.Annotations["squidflow.github.io/created-by"]
		app.ApplicationRuntime.LastUpdatedAt = time.Now() // TODO: fix this
		app.ApplicationRuntime.LastUpdatedBy = applicationRuntime.Annotations["squidflow.github.io/last-modified-by"]
	}

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get application detail: %v", err)})
		return
	}

	c.JSON(200, app)
}

func ApplicationsList(c *gin.Context) {
	tenant := c.GetString(middleware.TenantKey)
	username := c.GetString(middleware.UserNameKey)

	log.G().Infof("tenant: %s, username: %s", tenant, username)

	var project = tenant

	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: false,
	}
	cloneOpts.Parse()

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	apps, err := repowriter.Repo().RunAppList(context.Background(), &types.AppListOptions{
		CloneOpts:    cloneOpts,
		ProjectName:  project,
		ArgoCDClient: argoClient,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list applications: %v", err)})
		return
	}

	// update application runtime status
	for i, app := range apps {
		var argocdappname = fmt.Sprintf("%s-%s", app.ApplicationInstantiation.TenantName, app.ApplicationInstantiation.ApplicationName)
		log.G().WithFields(log.Fields{
			"application namespace": app.ApplicationInstantiation.TenantName,
			"application name":      store.Default.ArgoCDNamespace,
		}).Debug("get application status")

		//TODO: opt with client-go list method
		argoApp, err := argoClient.Applications(store.Default.ArgoCDNamespace).
			Get(context.Background(), argocdappname, metav1.GetOptions{})
		if err != nil {
			log.G().WithError(err).Error("Failed to get application")
			continue
		} else {
			apps[i].ApplicationRuntime.Status = getAppStatus(argoApp)
			apps[i].ApplicationRuntime.Health = getAppHealth(argoApp)
			apps[i].ApplicationRuntime.SyncStatus = getAppSyncStatus(argoApp)
			apps[i].ApplicationRuntime.ArgoCDUrl = fmt.Sprintf("https://argocd.squidflow.io/applications/%s", argocdappname)
			apps[i].ApplicationRuntime.CreatedAt = argoApp.CreationTimestamp.Time
			apps[i].ApplicationRuntime.CreatedBy = argoApp.Annotations["squidflow.github.io/created-by"]
			apps[i].ApplicationRuntime.LastUpdatedAt = time.Now() // TODO: fix this
			apps[i].ApplicationRuntime.LastUpdatedBy = argoApp.Annotations["squidflow.github.io/last-modified-by"]
		}
	}

	c.JSON(200, types.ApplicationListResponse{
		Total:   int64(len(apps)),
		Success: true,
		Message: "applications listed successfully",
		Items:   apps,
		Error:   "",
	})
}

// ApplicationSync handles the synchronization of one or more Argo CD applications
func ApplicationSync(c *gin.Context) {
	var req types.SyncApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
		return
	}

	// Create ArgoCD client
	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	response := types.SyncApplicationResponse{
		Results: make([]types.SyncApplicationResult, 0, len(req.Applications)),
	}

	// Process each application
	for _, appName := range req.Applications {
		result := types.SyncApplicationResult{
			Name: appName,
		}

		// Get the application
		app, err := argoClient.Applications(appName).Get(context.Background(), appName, metav1.GetOptions{})
		if err != nil {
			result.Status = "Failed"
			result.Message = fmt.Sprintf("Failed to get application: %v", err)
			response.Results = append(response.Results, result)
			continue
		}

		// Prepare sync operation
		syncOp := argocdv1alpha1.SyncOperation{
			Prune: false,
		}

		// Update application with sync operation
		app.Operation = &argocdv1alpha1.Operation{
			Sync: &syncOp,
		}

		// Apply the sync operation
		if err != nil {
			result.Status = "Failed"
			result.Message = fmt.Sprintf("Failed to sync application: %v", err)
		} else {
			result.Status = "Syncing"
			result.Message = "Application sync initiated successfully"
		}

		response.Results = append(response.Results, result)
		log.G().WithFields(log.Fields{
			"application": appName,
			"status":      result.Status,
			"message":     result.Message,
		}).Info("Application sync result")
	}

	c.JSON(200, response)
}

func ApplicationUpdate(c *gin.Context) {
	username := c.GetString(middleware.UserNameKey)
	tenant := c.GetString(middleware.TenantKey)
	appName := c.Param("name")

	log.G().WithFields(log.Fields{
		"username": username,
		"tenant":   tenant,
		"appName":  appName,
	}).Debug("update argo application")

	var updateReq types.ApplicationUpdateRequest
	if err := c.BindJSON(&updateReq); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	cloneOpts := &git.CloneOptions{
		Repo:     viper.GetString("application_repo.remote_url"),
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: true,
	}
	cloneOpts.Parse()

	annotations := make(map[string]string)
	if updateReq.ApplicationInstantiation.Description != "" {
		annotations["squidflow.github.io/description"] = updateReq.ApplicationInstantiation.Description
	}

	// TODO: support security
	log.G().WithFields(log.Fields{
		"security": updateReq.ApplicationInstantiation.Security,
	}).Debug("TODO support security")

	// TODO: support ingress
	log.G().WithFields(log.Fields{
		"ingress": updateReq.ApplicationInstantiation.Ingress,
	}).Debug("TODO support ingress")

	annotations[argocd.AnnotationKeyLastModifiedBy] = username
	annotations[argocd.AnnotationKeyLastModifiedAt] = time.Now().Format(time.RFC3339)

	updateOpts := &types.UpdateOptions{
		CloneOpts:   cloneOpts,
		ProjectName: tenant,
		AppName:     appName,
		Username:    username,
		UpdateReq:   &updateReq,
		KubeFactory: kube.NewFactory(),
		Annotations: annotations,
	}

	if err := repowriter.Repo().RunAppUpdate(context.Background(), updateOpts); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update application: %v", err)})
		return
	}

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	app, err := repowriter.Repo().RunAppGet(context.Background(), &types.AppListOptions{
		CloneOpts:    cloneOpts,
		ProjectName:  tenant,
		ArgoCDClient: argoClient,
	}, appName)

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get updated application details: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message":     "application updated successfully",
		"application": app,
	})
}

// ApplicationSourceValidate handles the request for validating application source
func ApplicationSourceValidate(c *gin.Context) {
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
	}).Info("starting application source validation")

	// Clone repository
	cloneOpts := &git.CloneOptions{
		Repo:          req.Repo,
		FS:            fs.Create(memfs.New()),
		CloneForWrite: false,
		Submodules:    req.Submodules,
	}
	cloneOpts.Parse()
	cloneOpts.SetRevision(req.TargetRevision)

	_, repofs, err := cloneOpts.GetRepo(context.Background())
	if err != nil {
		log.G().WithError(err).Error("failed to clone repository")
		c.JSON(400, gin.H{
			"success": false,
			"message": fmt.Sprintf("failed to clone repository: %v", err),
		})
		return
	}

	// Detect application type and validate structure
	appType, environments, err := reporeader.ValidateApplicationStructure(repofs, req)
	if err != nil {
		log.G().WithError(err).Error("failed to validate application structure")
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
	}).Info("detected application structure")

	memFS := memfs.New()
	suiteableEnv := []types.AppSourceWithEnvironment{}

	for _, env := range environments {
		log.G().WithFields(log.Fields{
			"type": appType,
			"env":  env,
		}).Debug("validating environment")

		envResult := types.AppSourceWithEnvironment{
			Environments: env,
			Manifest:     "",
			Valid:        true,
			Error:        "",
		}

		// generate manifest
		var manifests []byte
		switch appType {
		case reporeader.SourceHelm, reporeader.SourceHelmMultiEnv:
			manifests, err = dryrun.GenerateHelmManifest(repofs, req, env, "application1", "default")
			if err != nil {
				log.G().WithError(err).Error("failed to generate helm manifest")
				envResult.Valid = false
				envResult.Error = err.Error()
			}

		case reporeader.SourceKustomize, reporeader.SourceKustomizeMultiEnv:
			manifests, err = dryrun.GenerateKustomizeManifest(repofs, req, env, "application1", "default")
			if err != nil {
				log.G().WithError(err).Error("failed to generate kustomize manifest")
				envResult.Valid = false
				envResult.Error = err.Error()
			}
		}

		if err != nil {
			log.G().WithError(err).WithField("env", env).Error("failed to generate manifest")
			envResult.Valid = false
			envResult.Error = err.Error()
		} else {
			envResult.Manifest = string(manifests)

			log.G().WithFields(log.Fields{
				"env": env,
			}).Debug("writing manifest to memory file system")

			// write manifest to memory file system
			manifestPath := fmt.Sprintf("/manifests/%s.yaml", env)
			if err := memFS.MkdirAll("/manifests", 0755); err != nil {
				log.G().WithError(err).Error("failed to create manifests directory")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

			f, err := memFS.Create(manifestPath)
			if err != nil {
				log.G().WithError(err).Error("failed to create manifest file")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

			if _, err := f.Write(manifests); err != nil {
				f.Close()
				log.G().WithError(err).Error("failed to write manifest")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}
			f.Close()

			// validate manifest with kubeconform
			// TODO: make this offline
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
				log.G().WithError(err).Error("failed to create validator")
				envResult.Valid = false
				envResult.Error = err.Error()
				continue
			}

			f, err = memFS.Open(manifestPath)
			if err != nil {
				log.G().WithError(err).Error("failed to open manifest for validation")
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
					}).Error("manifest validation failed")
					break
				}
			}
		}

		suiteableEnv = append(suiteableEnv, envResult)
	}

	var allValid = true
	for _, env := range suiteableEnv {
		if !env.Valid {
			allValid = false
			break
		}
	}

	if allValid {
		c.JSON(200, types.ValidateAppSourceResponse{
			Success:      true,
			Message:      fmt.Sprintf("valid %s application source", appType),
			Type:         string(appType),
			SuiteableEnv: suiteableEnv,
		})
	} else {
		c.JSON(400, types.ValidateAppSourceResponse{
			Success:      false,
			Message:      fmt.Sprintf("valid %s application source", appType),
			Type:         string(appType),
			SuiteableEnv: suiteableEnv,
		})
	}
}
