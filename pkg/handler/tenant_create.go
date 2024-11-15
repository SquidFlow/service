package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
)

type ProjectCreateRequest struct {
	ProjectName string            `json:"project-name" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func CreateTenant(c *gin.Context) {
	var req ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
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

	opts := &ProjectCreateOptions{
		CloneOpts:   cloneOpts,
		ProjectName: req.ProjectName,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	}

	err := RunProjectCreate(context.Background(), opts)
	if err != nil {
		log.G().Errorf("Failed to create project: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create project: %v", err)})
		return
	}

	c.JSON(201, gin.H{
		"message": fmt.Sprintf("Project '%s' created successfully", req.ProjectName),
		"project": req,
	})
}

func RunProjectCreate(ctx context.Context, opts *ProjectCreateOptions) error {
	var (
		err                   error
		installationNamespace string
	)

	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, "")
	if err != nil {
		return err
	}

	installationNamespace, err = getInstallationNamespace(repofs)
	if err != nil {
		return fmt.Errorf(util.Doc("Bootstrap folder not found, please execute `<BIN> repo bootstrap --installation-path %s` command"), repofs.Root())
	}

	projectExists := repofs.ExistsOrDie(repofs.Join(store.Default.ProjectsDir, opts.ProjectName+".yaml"))
	if projectExists {
		return fmt.Errorf("project '%s' already exists", opts.ProjectName)
	}

	log.G().Debug("repository is ok")

	if opts.DestKubeServer == "" {
		opts.DestKubeServer = store.Default.DestServer
		if opts.DestKubeContext != "" {
			opts.DestKubeServer, err = util.KubeContextToServer(opts.DestKubeContext)
			if err != nil {
				return err
			}
		}
	}

	projectYAML, appsetYAML, clusterResReadme, clusterResConf, err := generateProjectManifests(&GenerateProjectOptions{
		Name:               opts.ProjectName,
		Namespace:          installationNamespace,
		RepoURL:            opts.CloneOpts.URL(),
		Revision:           opts.CloneOpts.Revision(),
		InstallationPath:   opts.CloneOpts.Path(),
		DefaultDestServer:  opts.DestKubeServer,
		DefaultDestContext: opts.DestKubeContext,
		Labels:             opts.Labels,
		Annotations:        opts.Annotations,
	})
	if err != nil {
		return fmt.Errorf("failed to generate project resources: %w", err)
	}

	if opts.DryRun {
		log.G().Printf("%s", util.JoinManifests(projectYAML, appsetYAML))
		return nil
	}

	bulkWrites := []fs.BulkWriteRequest{}

	if opts.DestKubeContext != "" {
		log.G().Infof("adding cluster: %s", opts.DestKubeContext)
		if err = opts.AddCmd.Execute(ctx, opts.DestKubeContext); err != nil {
			return fmt.Errorf("failed to add new cluster credentials: %w", err)
		}

		if !repofs.ExistsOrDie(repofs.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, opts.DestKubeContext)) {
			bulkWrites = append(bulkWrites, fs.BulkWriteRequest{
				Filename: repofs.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, opts.DestKubeContext+".json"),
				Data:     clusterResConf,
				ErrMsg:   "failed to write cluster config",
			})

			bulkWrites = append(bulkWrites, fs.BulkWriteRequest{
				Filename: repofs.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, opts.DestKubeContext, "README.md"),
				Data:     clusterResReadme,
				ErrMsg:   "failed to write cluster resources readme",
			})
		}
	}

	bulkWrites = append(bulkWrites, fs.BulkWriteRequest{
		Filename: repofs.Join(store.Default.ProjectsDir, opts.ProjectName+".yaml"),
		Data:     util.JoinManifests(projectYAML, appsetYAML),
		ErrMsg:   "failed to create project file",
	})

	if err = fs.BulkWrite(repofs, bulkWrites...); err != nil {
		return err
	}

	log.G().Infof("pushing new project manifest to repo")
	if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: fmt.Sprintf("chore: added project '%s'", opts.ProjectName)}); err != nil {
		return err
	}

	log.G().Infof("project created: '%s'", opts.ProjectName)

	return nil
}
