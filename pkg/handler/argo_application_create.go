package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/middleware"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
)

type (
	AppCreateOptions struct {
		CloneOpts       *git.CloneOptions
		AppsCloneOpts   *git.CloneOptions
		ProjectName     string
		KubeContextName string
		createOpts      *application.CreateOptions
		KubeFactory     kube.Factory
		Timeout         time.Duration
		Labels          map[string]string
		Annotations     map[string]string
		Include         string
		Exclude         string
	}

	DestinationClusters struct {
		Clusters  []string `json:"clusters"`
		Namespace string   `json:"namespace"`
	}

	TLS struct {
		Enabled    bool   `json:"enabled"`
		SecretName string `json:"secretName"`
	}

	Ingress struct {
		Host string `json:"host"`
		TLS  *TLS   `json:"tls,omitempty"`
	}

	SecretStoreRef struct {
		ID string `json:"id"`
	}

	ExternalSecret struct {
		SecretStoreRef SecretStoreRef `json:"secret_store_ref"`
	}

	Security struct {
		ExternalSecret *ExternalSecret `json:"external_secret,omitempty"`
	}

	ApplicationCreate struct {
		ApplicationSource   ApplicationSource   `json:"application_source"`
		ApplicationName     string              `json:"application_name"`
		TenantName          string              `json:"tenant_name"`
		AppCode             string              `json:"appcode"`
		Description         string              `json:"description"`
		DestinationClusters DestinationClusters `json:"destination_clusters"`
		Ingress             *Ingress            `json:"ingress,omitempty"`
		Security            *Security           `json:"security,omitempty"`
		IsDryRun            bool                `json:"is_dryrun"`
	}

	ApplicationSource struct {
		Type           string `json:"type" binding:"required,oneof=git"`
		URL            string `json:"url" binding:"required"`
		TargetRevision string `json:"targetRevision" binding:"required"`
		Path           string `json:"path" binding:"required"`
	}
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

func CreateArgoApplication(c *gin.Context) {
	username := c.GetString(middleware.UserNameKey)
	tenant := c.GetString(middleware.TenantKey)
	log.G().WithFields(log.Fields{
		"username": username,
		"tenant":   tenant,
	}).Debug("create argo application")

	var createReq ApplicationCreate
	if err := c.BindJSON(&createReq); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if tenant != createReq.TenantName {
		c.JSON(400, gin.H{"error": "tenant in request body does not match tenant in authorization header"})
		return
	}

	if err := validateApplication(&createReq); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// TODO: dry run
	if createReq.IsDryRun {
		c.JSON(200, gin.H{
			"message":     "Validation passed",
			"application": createReq,
		})
		return
	}

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
			AppName:          createReq.ApplicationName,
			AppType:          application.AppTypeKustomize,
			AppSpecifier:     createReq.ApplicationSource.URL,
			InstallationMode: application.InstallationModeNormal,
			DestServer:       "https://kubernetes.default.svc",
			Annotations: map[string]string{
				"h4-poc.github.io/created-by":  username,
				"h4-poc.github.io/tenant":      tenant,
				"h4-poc.github.io/description": createReq.Description,
				"h4-poc.github.io/appcode":     createReq.AppCode,
			},
		},
		ProjectName: createReq.TenantName,
		KubeFactory: kube.NewFactory(),
	}
	opt.CloneOpts.Parse()
	opt.AppsCloneOpts.Parse()

	if createReq.Ingress != nil {
		opt.createOpts.Annotations["ingress.host"] = createReq.Ingress.Host
		if createReq.Ingress.TLS != nil {
			opt.createOpts.Annotations["ingress.tls.enabled"] = fmt.Sprintf("%v", createReq.Ingress.TLS.Enabled)
			opt.createOpts.Annotations["ingress.tls.secretName"] = createReq.Ingress.TLS.SecretName
		}
	}

	if createReq.Security != nil && createReq.Security.ExternalSecret != nil {
		opt.createOpts.Annotations["security.external-secret.store-id"] = createReq.Security.ExternalSecret.SecretStoreRef.ID
	}

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

func validateApplication(app *ApplicationCreate) error {
	if app.ApplicationName == "" {
		return fmt.Errorf("application_name is required")
	}
	if app.TenantName == "" {
		return fmt.Errorf("tenant_name is required")
	}
	if app.Description == "" {
		return fmt.Errorf("description is required")
	}

	if err := validateApplicationSource(&app.ApplicationSource); err != nil {
		return fmt.Errorf("invalid application_source: %w", err)
	}

	if err := validateDestinationClusters(&app.DestinationClusters); err != nil {
		return fmt.Errorf("invalid destination_cluster: %w", err)
	}

	if app.Ingress != nil {
		if err := validateIngress(app.Ingress); err != nil {
			return fmt.Errorf("invalid ingress configuration: %w", err)
		}
	}

	if app.Security != nil {
		if err := validateSecurity(app.Security); err != nil {
			return fmt.Errorf("invalid security configuration: %w", err)
		}
	}

	return nil
}

func validateApplicationSource(source *ApplicationSource) error {
	if source == nil {
		return fmt.Errorf("application_source is required")
	}
	if source.Type != "git" {
		return fmt.Errorf("source type must be git")
	}
	if source.URL == "" {
		return fmt.Errorf("source URL is required")
	}
	if source.TargetRevision == "" {
		return fmt.Errorf("source targetRevision is required")
	}
	if source.Path == "" {
		return fmt.Errorf("source path is required")
	}
	return nil
}

func validateDestinationClusters(dest *DestinationClusters) error {
	if dest == nil {
		return fmt.Errorf("destination_clusters is required")
	}
	if len(dest.Clusters) == 0 {
		return fmt.Errorf("at least one destination cluster must be specified")
	}
	if dest.Namespace == "" {
		dest.Namespace = "default"
	}
	return nil
}

func validateIngress(ingress *Ingress) error {
	if ingress.Host == "" {
		return fmt.Errorf("ingress host is required when ingress is enabled")
	}
	if ingress.TLS != nil {
		if ingress.TLS.Enabled && ingress.TLS.SecretName == "" {
			return fmt.Errorf("TLS secret name is required when TLS is enabled")
		}
	}
	return nil
}

func validateSecurity(security *Security) error {
	if security.ExternalSecret != nil {
		if security.ExternalSecret.SecretStoreRef.ID == "" {
			return fmt.Errorf("secret store ID is required when external secret is enabled")
		}
	}
	return nil
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
