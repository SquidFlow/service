package handler

import (
	"context"
	"fmt"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	repowriter "github.com/squidflow/service/pkg/repo/writer"
	"github.com/squidflow/service/pkg/source"
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
		c.JSON(400, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	if tenant != createReq.ApplicationInstantiation.TenantName {
		c.JSON(400, gin.H{
			"error": "ApplicationInstantiation field tenant in request body does not match tenant in authorization header",
		})
		return
	}

	// check the application source is valid add it to cache
	appCloneOpts := &git.CloneOptions{
		Repo: application.BuildKustomizeResourceRef(application.ApplicationSourceOption{
			Repo:           createReq.ApplicationSource.Repo,
			Path:           createReq.ApplicationSource.Path,
			TargetRevision: createReq.ApplicationSource.TargetRevision,
		}),
		FS:            fs.Create(memfs.New()),
		CloneForWrite: false,
		Submodules:    true,
	}
	appCloneOpts.Parse()
	_, appfs, err := appCloneOpts.GetRepo(context.Background())
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("failed to clone application source repository: %v", err)})
		return
	}

	appSource, err := source.NewAppSource(
		appfs,
		createReq.ApplicationSource.Path,
		createReq.ApplicationSource.ApplicationSpecifier.HelmManifestPath,
	)
	if err != nil {
		log.G().WithError(err).Error("failed to create app source")
		c.JSON(400, gin.H{"error": fmt.Sprintf("failed to create app source: %v", err)})
		return
	}

	// Normal application creation flow
	var opt = application.AppCreateOptions{
		AppOpts: &application.CreateOptions{
			AppName: createReq.ApplicationInstantiation.ApplicationName,
			AppType: appSource.GetType(),
			AppSpecifier: application.BuildKustomizeResourceRef(application.ApplicationSourceOption{
				Repo:           createReq.ApplicationSource.Repo,
				Path:           createReq.ApplicationSource.Path,
				TargetRevision: createReq.ApplicationSource.TargetRevision,
			}),
			InstallationMode: application.InstallModeType(createReq.ApplicationInstantiation.InstallationMode),
			DestServer:       "https://kubernetes.default.svc",
			Annotations: map[string]string{
				argocd.AnnotationKeyEnvironment: username,
				argocd.AnnotationKeyTenant:      tenant,
				argocd.AnnotationKeyDescription: createReq.ApplicationInstantiation.Description,
				argocd.AnnotationKeyAppCode:     createReq.ApplicationInstantiation.AppCode,
			},
			AppSource: appSource,
		},
		ProjectName: createReq.ApplicationInstantiation.TenantName,
		KubeFactory: kube.NewFactory(),
		DryRun:      createReq.IsDryRun,
	}

	log.G().WithFields(log.Fields{
		"appOpts": opt.AppOpts,
	}).Debug("create application options: ")

	// TODO: support multiple clusters
	createResp, err := repowriter.TenantRepo(tenant).RunAppCreate(context.Background(), &opt)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("failed to create application in cluster %s: %v", opt.AppOpts.DestServer, err)})
		return
	}

	c.JSON(201, createResp)
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

	// 1. delete from gitops repo first
	if err := repowriter.TenantRepo(tenant).RunAppDelete(context.Background(), appName); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete application: %v", err)})
		return
	}

	// 2. delete from kubernetes
	// pull request mode, do not delete from kubernetes
	if viper.GetString("gitops.mode") != "pull_request" {
		go func(projectName string, appName string) {
			argoClient, err := kube.NewArgoCdClient()
			if err != nil {
				log.G().WithFields(log.Fields{
					"projectName": projectName,
					"appName":     appName,
				}).Warn("delete application failed to create ArgoCD client")
				return
			}
			applicationName := fmt.Sprintf("%s-%s", projectName, appName)
			err = argoClient.Applications(store.Default.ArgoCDNamespace).Delete(context.Background(), applicationName, metav1.DeleteOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					log.G().WithFields(log.Fields{
						"projectName": projectName,
						"appName":     appName,
					}).Warn("delete application handler: application not found")
				} else {
					log.G().WithFields(log.Fields{
						"projectName": projectName,
						"appName":     appName,
					}).Warn("delete application handler: failed to delete application")
				}
			}
		}(tenant, appName)
	}

	c.JSON(204, nil)
}

func ApplicationGet(c *gin.Context) {
	tenant := c.GetString(middleware.TenantKey)
	username := c.GetString(middleware.UserNameKey)
	appName := c.Param("name")

	log.G().Infof("tenant: %s, username: %s, appName: %s", tenant, username, appName)

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	app, err := repowriter.TenantRepo(tenant).RunAppGet(context.Background(), appName)

	var argocdappname = fmt.Sprintf("%s-%s", app.ApplicationInstantiation.TenantName, app.ApplicationInstantiation.ApplicationName)
	log.G().WithFields(log.Fields{
		"application namespace": app.ApplicationInstantiation.TenantName,
		"application name":      store.Default.ArgoCDNamespace,
	}).Debug("get application status")

	//TODO: opt with list method
	applicationRuntime, err := argoClient.Applications(store.Default.ArgoCDNamespace).
		Get(context.Background(), argocdappname, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.G().WithFields(log.Fields{
				"application": appName,
				"namespace":   store.Default.ArgoCDNamespace,
			}).Info("application not install in argocd")
		} else {
			log.G().WithError(err).Error("failed to get application")
		}
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
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get application detail: %v", err)})
		return
	}

	c.JSON(200, app)
}

func ApplicationsList(c *gin.Context) {
	tenant := c.GetString(middleware.TenantKey)
	username := c.GetString(middleware.UserNameKey)

	log.G().Infof("tenant: %s, username: %s", tenant, username)

	argoClient, err := kube.NewArgoCdClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
		return
	}

	apps, err := repowriter.TenantRepo(tenant).RunAppList(context.Background())
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("failed to list applications: %v", err)})
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

// TODO: not implemented
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
		ProjectName: tenant,
		AppName:     appName,
		Username:    username,
		UpdateReq:   &updateReq,
		KubeFactory: kube.NewFactory(),
		Annotations: annotations,
	}

	if err := repowriter.TenantRepo(tenant).RunAppUpdate(context.Background(), updateOpts); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update application: %v", err)})
		return
	}

	// TODO: update application in ArgoCD
	//argoClient, err := kube.NewArgoCdClient()
	//if err != nil {
	//	c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create ArgoCD client: %v", err)})
	//	return
	//}

	app, err := repowriter.TenantRepo(tenant).RunAppGet(context.Background(), appName)

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

	// Create appropriate AppSource based on the repository content
	appSource, err := source.NewAppSource(repofs, req.Path, req.ApplicationSpecifier.HelmManifestPath)
	if err != nil {
		log.G().WithError(err).Error("failed to create app source")
		c.JSON(400, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate each environment and generate manifests
	suiteableEnv := []types.AppSourceWithEnvironment{}
	allValid := true

	for _, env := range appSource.DetectEnvironments() {
		envResult := types.AppSourceWithEnvironment{
			Environments: env,
			Valid:        true,
		}

		// Generate manifest
		manifest, err := appSource.Manifest(env)
		if err != nil {
			log.G().WithError(err).WithFields(log.Fields{
				"env": env,
			}).Error("failed to generate manifest")
			envResult.Valid = false
			envResult.Error = err.Error()
			allValid = false
		} else {
			envResult.Manifest = string(manifest)
		}

		suiteableEnv = append(suiteableEnv, envResult)
	}

	// Validate all environments
	validationResults := appSource.Validate(repofs, req.Path)
	for env, err := range validationResults {
		for i, result := range suiteableEnv {
			if result.Environments == env && err != nil {
				suiteableEnv[i].Valid = false
				suiteableEnv[i].Error = err.Error()
				allValid = false
				break
			}
		}
	}

	if allValid {
		c.JSON(200, types.ValidateAppSourceResponse{
			Success:      true,
			Message:      fmt.Sprintf("valid %s application source", appSource.GetType()),
			Type:         appSource.GetType(),
			SuiteableEnv: suiteableEnv,
		})
	} else {
		c.JSON(400, types.ValidateAppSourceResponse{
			Success:      false,
			Message:      fmt.Sprintf("invalid %s application source", appSource.GetType()),
			Type:         appSource.GetType(),
			SuiteableEnv: suiteableEnv,
		})
	}
}
