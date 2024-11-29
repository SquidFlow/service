package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
)

type AppDeleteOptions struct {
	CloneOpts   *git.CloneOptions
	ProjectName string
	AppName     string
	Global      bool
}

func DeleteArgoApplication(c *gin.Context) {
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

	if err := deleteApplication(context.Background(), &AppDeleteOptions{
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

func deleteApplication(ctx context.Context, opts *AppDeleteOptions) error {
	r, repofs, err := prepareRepo(ctx, opts.CloneOpts, opts.ProjectName)
	if err != nil {
		return err
	}

	appDir := repofs.Join(store.Default.AppsDir, opts.AppName)
	appExists := repofs.ExistsOrDie(appDir)
	if !appExists {
		return fmt.Errorf("application '%s' not found", opts.AppName)
	}

	var dirToRemove string
	commitMsg := fmt.Sprintf("Deleted app '%s'", opts.AppName)
	if opts.Global {
		dirToRemove = appDir
	} else {
		appOverlaysDir := repofs.Join(appDir, store.Default.OverlaysDir)
		overlaysExists := repofs.ExistsOrDie(appOverlaysDir)
		if !overlaysExists {
			appOverlaysDir = appDir
		}

		appProjectDir := repofs.Join(appOverlaysDir, opts.ProjectName)
		overlayExists := repofs.ExistsOrDie(appProjectDir)
		if !overlayExists {
			return fmt.Errorf("application '%s' not found in project '%s'", opts.AppName, opts.ProjectName)
		}

		allOverlays, err := repofs.ReadDir(appOverlaysDir)
		if err != nil {
			return fmt.Errorf("failed to read overlays directory '%s': %w", appOverlaysDir, err)
		}

		if len(allOverlays) == 1 {
			dirToRemove = appDir
		} else {
			commitMsg += fmt.Sprintf(" from project '%s'", opts.ProjectName)
			dirToRemove = appProjectDir
		}
	}

	err = billyUtils.RemoveAll(repofs, dirToRemove)
	if err != nil {
		return fmt.Errorf("failed to delete directory '%s': %w", dirToRemove, err)
	}

	log.G().Info("committing changes to gitops repo...")
	if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: commitMsg}); err != nil {
		return fmt.Errorf("failed to push to repo: %w", err)
	}

	return nil
}
