package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
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

	var createReq ApplicationCreateRequest
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
	var opt = AppCreateOptions{
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
		createOpts: &application.CreateOptions{
			AppName:          createReq.ApplicationInstantiation.ApplicationName,
			AppType:          application.AppTypeKustomize,
			AppSpecifier:     buildKustomizeResourceRef(createReq.ApplicationSource),
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

	if err := RunAppCreate(context.Background(), &opt); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to create application in cluster %s: %v", opt.createOpts.DestServer, err)})
		return
	}

	c.JSON(201, gin.H{
		"message":     "Applications created successfully",
		"application": createReq,
	})
}

func RunAppCreate(ctx context.Context, opts *AppCreateOptions) error {
	var (
		appsRepo git.Repository
		appsfs   fs.FS
	)

	log.G().WithFields(log.Fields{
		"app-url":      opts.AppsCloneOpts.URL(),
		"app-revision": opts.AppsCloneOpts.Revision(),
		"app-path":     opts.AppsCloneOpts.Path(),
	}).Debug("starting with options: ")

	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return err
	}
	log.G().Debugf("repofs: %v", repofs)

	if opts.AppsCloneOpts.Repo != "" {
		if opts.AppsCloneOpts.Auth.Password == "" {
			opts.AppsCloneOpts.Auth.Username = opts.CloneOpts.Auth.Username
			opts.AppsCloneOpts.Auth.Password = opts.CloneOpts.Auth.Password
			opts.AppsCloneOpts.Auth.CertFile = opts.CloneOpts.Auth.CertFile
			opts.AppsCloneOpts.Provider = opts.CloneOpts.Provider
		}

		appsRepo, appsfs, err = getRepo(ctx, opts.AppsCloneOpts)
		if err != nil {
			return err
		}
	} else {
		opts.AppsCloneOpts = opts.CloneOpts
		appsRepo, appsfs = r, repofs
	}

	if err = setAppOptsDefaults(ctx, repofs, opts); err != nil {
		return err
	}

	app, err := parseApp(opts.createOpts, opts.ProjectName, opts.CloneOpts.URL(), opts.CloneOpts.Revision(), opts.CloneOpts.Path())
	if err != nil {
		return fmt.Errorf("failed to parse application from flags: %w", err)
	}

	if err = app.CreateFiles(repofs, appsfs, opts.ProjectName); err != nil {
		if errors.Is(err, application.ErrAppAlreadyInstalledOnProject) {
			return fmt.Errorf("application '%s' already exists in project '%s': %w", app.Name(), opts.ProjectName, err)
		}

		return err
	}

	if opts.AppsCloneOpts != opts.CloneOpts {
		log.G().Info("committing changes to apps repo...")
		if _, err = appsRepo.Persist(ctx, &git.PushOptions{
			CommitMsg: genCommitMsg("chore: "+ActionTypeCreate, ResourceNameApp, opts.createOpts.AppName, opts.ProjectName, repofs),
		}); err != nil {
			return fmt.Errorf("failed to push to apps repo: %w", err)
		}
	}

	log.G().Info("committing changes to git-ops repo...")
	var opt = git.PushOptions{CommitMsg: genCommitMsg("chore: "+ActionTypeCreate, ResourceNameApp, opts.createOpts.AppName, opts.ProjectName, repofs)}
	log.G().Debugf("git push option: %v", opt)
	revision, err := r.Persist(ctx, &opt)
	if err != nil {
		return fmt.Errorf("failed to push to gitops repo: %w", err)
	}

	log.G().Infof("installed application: %s and revision: %s", opts.createOpts.AppName, revision)
	return nil
}

func performDryRun(ctx context.Context, req *ApplicationCreateRequest) (*ApplicationDryRunResult, error) {
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
	appType, environments, err := validateApplicationStructure(repofs, req.ApplicationSource)
	if err != nil {
		return nil, err
	}

	log.G().WithFields(log.Fields{
		"type":         appType,
		"environments": environments,
	}).Debug("Detected application structure")

	// Initialize dry run result
	result := &ApplicationDryRunResult{
		Success:      true,
		Total:        len(environments),
		Environments: make([]ApplicationDryRunEnv, 0, len(environments)),
	}

	// For each environment, render and validate the templates
	for _, env := range environments {
		log.G().WithFields(log.Fields{
			"environment": env,
			"type":        appType,
		}).Debug("Processing environment")

		envResult := ApplicationDryRunEnv{
			Environment: env,
			IsValid:     true,
		}

		var manifests []byte
		switch appType {
		case "helm":
			manifests, err = generateHelmManifest(repofs, &req.ApplicationSource, env, req.ApplicationInstantiation.ApplicationName, req.ApplicationTarget[0].Namespace)
		case "kustomize":
			manifests, err = generateKustomizeManifest(repofs, &req.ApplicationSource, env, req.ApplicationInstantiation.ApplicationName, req.ApplicationTarget[0].Namespace)
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

// generateHelmManifest generates Helm manifests for a specific environment
func generateHelmManifest(repofs fs.FS, req *ApplicationSourceRequest, env string, applicationName string, applicationNamespace string) ([]byte, error) {
	log.G().WithFields(log.Fields{
		"path":      req.Path,
		"env":       env,
		"name":      applicationName,
		"namespace": applicationNamespace,
	}).Debug("Preparing helm template")

	// determine chart path
	var chartPath string
	if req.ApplicationSpecifier != nil && req.ApplicationSpecifier.HelmManifestPath != "" {
		// case 1: use specified helm manifest path
		chartPath = repofs.Join(req.Path, req.ApplicationSpecifier.HelmManifestPath)
	} else {
		// case 2: directly use the specified path to find Chart.yaml
		chartPath = req.Path
	}

	log.G().WithFields(log.Fields{
		"chartPath": chartPath,
	}).Debug("Looking for chart")

	// validate Chart.yaml exists
	if !repofs.ExistsOrDie(repofs.Join(chartPath, "Chart.yaml")) {
		return nil, fmt.Errorf("Chart.yaml not found at path: %s", chartPath)
	}

	// create temp directory for chart files
	tmpDir, err := os.MkdirTemp("", "helm-chart-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// copy chart files to temp directory
	if err := copyChartFiles(repofs, chartPath, tmpDir); err != nil {
		return nil, fmt.Errorf("failed to copy chart files: %w", err)
	}

	// read values file
	var valuesContent []byte
	if env != "default" {
		// check if environment specific values directory exists
		envValuesPath := repofs.Join(req.Path, "environments", env, "values.yaml")
		if repofs.ExistsOrDie(envValuesPath) {
			log.G().WithFields(log.Fields{
				"valuesPath": envValuesPath,
			}).Debug("Reading environment values")

			valuesContent, err = repofs.ReadFile(envValuesPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read values file for environment %s: %w", env, err)
			}
		} else {
			// if no environment specific values, use default
			valuesPath := repofs.Join(chartPath, "values.yaml")
			log.G().WithFields(log.Fields{
				"valuesPath": valuesPath,
			}).Debug("Environment values not found, using default values")

			valuesContent, err = repofs.ReadFile(valuesPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read values file: %w", err)
			}
		}
	} else {
		// use default values.yaml
		valuesPath := repofs.Join(chartPath, "values.yaml")
		log.G().WithFields(log.Fields{
			"valuesPath": valuesPath,
		}).Debug("Reading default values")

		valuesContent, err = repofs.ReadFile(valuesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file: %w", err)
		}
	}

	// parse values
	values := map[string]interface{}{}
	if err := yaml.Unmarshal(valuesContent, &values); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %w", err)
	}

	// create action configuration
	settings := cli.New()
	actionConfig := new(action.Configuration)

	// init action configuration
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		applicationName,
		"secrets",
		log.G().Debugf,
	); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}

	// create install action and configure dry run
	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.ReleaseName = applicationName
	client.Namespace = applicationNamespace
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
		"env":       env,
		"chartPath": chartPath,
		"namespace": applicationNamespace,
	}).Debug("Successfully rendered helm templates")

	return []byte(rel.Manifest), nil
}

// generateKustomizeManifest generates Kustomize manifests for a specific environment
func generateKustomizeManifest(repofs fs.FS, req *ApplicationSourceRequest, env string, applicationName string, applicationNamespace string) ([]byte, error) {
	log.G().WithFields(log.Fields{
		"path":      req.Path,
		"env":       env,
		"name":      applicationName,
		"namespace": applicationNamespace,
	}).Debug("Preparing kustomize build")

	// Create an in-memory filesystem for kustomize
	memFS := filesys.MakeFsInMemory()

	// configure build path
	var buildPath string
	if env == "default" {
		// case 1: simple Kustomize, directly use the specified path
		buildPath = req.Path

		// check if kustomization.yaml exists
		if !repofs.ExistsOrDie(repofs.Join(buildPath, "kustomization.yaml")) {
			return nil, fmt.Errorf("kustomization.yaml not found in %s", buildPath)
		}

		// copy the whole directory to memory filesystem
		err := copyToMemFS(repofs, buildPath, "/", memFS)
		if err != nil {
			return nil, fmt.Errorf("failed to copy files: %w", err)
		}
		buildPath = "/"
	} else {
		// case 2: multi-environment Kustomize, use overlays structure
		overlayPath := repofs.Join(req.Path, "overlays", env)
		if !repofs.ExistsOrDie(overlayPath) {
			return nil, fmt.Errorf("overlay directory for environment %s not found", env)
		}

		// check if kustomization.yaml exists in overlay
		if !repofs.ExistsOrDie(repofs.Join(overlayPath, "kustomization.yaml")) {
			return nil, fmt.Errorf("kustomization.yaml not found in overlay %s", env)
		}

		// copy the whole application directory (including base and overlays) to memory filesystem
		err := copyToMemFS(repofs, req.Path, "/", memFS)
		if err != nil {
			return nil, fmt.Errorf("failed to copy files: %w", err)
		}
		buildPath = repofs.Join("/overlays", env)
	}

	// List files for debugging
	entries, err := memFS.ReadDir(buildPath)
	if err == nil {
		fileNames := make([]string, 0)
		for _, entry := range entries {
			fileNames = append(fileNames, entry)
		}
		log.G().WithFields(log.Fields{
			"path":  buildPath,
			"files": fileNames,
		}).Debug("Files in memory filesystem")
	}

	// Create kustomize build options
	opts := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(opts)

	// Build manifests using the in-memory filesystem
	m, err := k.Run(memFS, buildPath)
	if err != nil {
		log.G().WithFields(log.Fields{
			"error": err,
			"path":  buildPath,
		}).Error("Failed to build kustomize")
		return nil, fmt.Errorf("failed to build kustomize: %w", err)
	}

	// Get YAML output
	yaml, err := m.AsYaml()
	if err != nil {
		return nil, fmt.Errorf("failed to generate yaml: %w", err)
	}

	log.G().Debug("Successfully generated kustomize manifest")
	return yaml, nil
}

// Helper function to copy files from repofs to memory filesystem
func copyToMemFS(repofs fs.FS, srcPath, destPath string, memFS filesys.FileSystem) error {
	entries, err := repofs.ReadDir(srcPath)
	if err != nil {
		return err
	}

	// Create the destination directory in memFS
	if err := memFS.MkdirAll(destPath); err != nil {
		return fmt.Errorf("failed to create directory %s in memory fs: %w", destPath, err)
	}

	for _, entry := range entries {
		srcFilePath := repofs.Join(srcPath, entry.Name())
		destFilePath := filepath.Join(destPath, entry.Name())

		if entry.IsDir() {
			if err := copyToMemFS(repofs, srcFilePath, destFilePath, memFS); err != nil {
				return err
			}
			continue
		}

		// Read file content from repofs
		content, err := repofs.ReadFile(srcFilePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", srcFilePath, err)
		}

		// Write file to memFS
		err = memFS.WriteFile(destFilePath, content)
		if err != nil {
			return fmt.Errorf("failed to write file %s to memory fs: %w", destFilePath, err)
		}
	}

	return nil
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

// buildKustomizeResourceRef builds a kustomize resource reference from an ApplicationSourceRequest
func buildKustomizeResourceRef(source ApplicationSourceRequest) string {
	// remove possible .git suffix
	repoURL := strings.TrimSuffix(source.Repo, ".git")

	// if git@ format, convert to https:// format
	if strings.HasPrefix(repoURL, "git@") {
		repoURL = strings.Replace(repoURL, "git@", "", 1)
		repoURL = strings.Replace(repoURL, ":", "/", 1)
	}

	// remove https:// prefix if exists
	repoURL = strings.TrimPrefix(repoURL, "https://")

	// build path part
	pathPart := ""
	if source.Path != "" {
		pathPart = "/" + source.Path
	}

	// build reference
	ref := source.TargetRevision
	if ref == "" {
		ref = "main" // default to main branch
	}

	// return format: repository/path?ref=revision
	return fmt.Sprintf("%s%s?ref=%s", repoURL, pathPart, ref)
}
