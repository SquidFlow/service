package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/spf13/viper"

	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
)

type (
	AppCreateOptions struct {
		CloneOpts       *git.CloneOptions // for ?
		AppsCloneOpts   *git.CloneOptions // for ?
		ProjectName     string
		KubeContextName string
		AppOpts         *application.CreateOptions // for ?
		KubeFactory     kube.Factory
		Timeout         time.Duration
		Labels          map[string]string
		Annotations     map[string]string
		Include         string
		Exclude         string
	}

	AppDeleteOptions struct {
		CloneOpts   *git.CloneOptions
		ProjectName string
		AppName     string
		Global      bool
	}

	AppListOptions struct {
		CloneOpts   *git.CloneOptions
		ProjectName string
	}
)

var (
	prepareRepo = func(ctx context.Context, cloneOpts *git.CloneOptions, projectName string) (git.Repository, fs.FS, error) {
		log.WithFields(log.Fields{
			"repoURL":  cloneOpts.URL(),
			"revision": cloneOpts.Revision(),
			"forWrite": cloneOpts.CloneForWrite,
		}).Debug("starting with options: ")

		// clone repo
		log.Infof("cloning git repository: %s", cloneOpts.URL())
		r, repofs, err := getRepo(ctx, cloneOpts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed cloning the repository: %w", err)
		}

		root := repofs.Root()
		log.Infof("using revision: \"%s\", installation path: \"%s\"", cloneOpts.Revision(), root)
		if !repofs.ExistsOrDie(store.Default.BootsrtrapDir) {
			return nil, nil, fmt.Errorf("bootstrap directory not found, please execute `repo bootstrap` command")
		}

		if projectName != "" {
			projExists := repofs.ExistsOrDie(repofs.Join(store.Default.ProjectsDir, projectName+".yaml"))
			if !projExists {
				return nil, nil, fmt.Errorf(util.Doc(fmt.Sprintf("project '%[1]s' not found, please execute `<BIN> project create %[1]s`", projectName)))
			}
		}

		log.Debug("repository is ok")

		return r, repofs, nil
	}

	getRepo = func(ctx context.Context, cloneOpts *git.CloneOptions) (git.Repository, fs.FS, error) {
		return cloneOpts.GetRepo(ctx)
	}
)

type Application struct {
	ProjectName string `json:"project-name"`
	AppName     string `json:"app-name"`
	Repo        string `json:"repo"`
	WaitTimeout string `json:"wait-timeout"`
}

func CreateApplication(c *gin.Context) {
	var createAppReq Application
	if err := c.BindJSON(&createAppReq); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	timeout, _ := time.ParseDuration(createAppReq.WaitTimeout)
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	var opt = AppCreateOptions{
		CloneOpts: &git.CloneOptions{
			Repo:     viper.GetString("application_repo.remote_url"),
			FS:       fs.Create(memfs.New()),
			Provider: "github",
			Auth: git.Auth{
				Password: viper.GetString("application_repo.access_token"),
			},
		},
		AppsCloneOpts: &git.CloneOptions{
			Repo: createAppReq.Repo,
			FS:   fs.Create(memfs.New()),
		},
		AppOpts: &application.CreateOptions{
			AppName:          createAppReq.AppName,
			AppType:          application.AppTypeDirectory,
			AppSpecifier:     "github.com/h4-poc/demo-app", // TODO
			InstallationMode: application.InstallationModeNormal,
			Labels:           nil,
			Annotations:      nil,
			Include:          "",
			Exclude:          "",
		},
		ProjectName: createAppReq.ProjectName,
		Timeout:     timeout,
		KubeFactory: kube.NewFactory(),
	}
	opt.CloneOpts.Parse()
	opt.AppsCloneOpts.Parse()

	if err := RunAppCreate(context.Background(), &opt); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create application: " + err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message":     "Application created successfully",
		"application": createAppReq,
	})
}

func RunAppCreate(ctx context.Context, opts *AppCreateOptions) error {
	var (
		appsRepo git.Repository
		appsfs   fs.FS
	)

	log.WithFields(log.Fields{
		"app-url":      opts.AppsCloneOpts.URL(),
		"app-revision": opts.AppsCloneOpts.Revision(),
		"app-path":     opts.AppsCloneOpts.Path(),
	}).Debug("starting with options: ")

	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		log.Errorf("failed to prepare gitops repo: %v", err)
		return err
	}

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

	app, err := parseApp(opts.AppOpts, opts.ProjectName, opts.CloneOpts.URL(), opts.CloneOpts.Revision(), opts.CloneOpts.Path())
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
		log.Info("committing changes to apps repo...")
		if _, err = appsRepo.Persist(ctx, &git.PushOptions{CommitMsg: getCommitMsg(opts, appsfs)}); err != nil {
			return fmt.Errorf("failed to push to apps repo: %w", err)
		}
	}

	log.Info("committing changes to gitops repo...")
	revision, err := r.Persist(ctx, &git.PushOptions{CommitMsg: getCommitMsg(opts, repofs)})
	if err != nil {
		return fmt.Errorf("failed to push to gitops repo: %w", err)
	}
	log.Debugf("pushed to gitops repo at revision: %s", revision)
	log.Infof("installed application: %s", opts.AppOpts.AppName)
	return nil
}
