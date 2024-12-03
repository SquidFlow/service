package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

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
			AppName:          createReq.ApplicationInstantiation.ApplicationName,
			AppType:          application.AppTypeKustomize,
			AppSpecifier:     createReq.ApplicationSource.Repo,
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