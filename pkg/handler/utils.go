package handler

import (
	"context"
	"fmt"
	"sync"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/application/repowriter"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

// getNewId returns a new id for the resource
func getNewId() string {
	return uuid.New().String()
}

// getAppStatus returns the status of the ArgoCD application
func getAppStatus(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}

	// Check if OperationState exists and has Phase
	if app.Status.OperationState != nil && app.Status.OperationState.Phase != "" {
		return string(app.Status.OperationState.Phase)
	}

	// If no OperationState, try to get status from Sync
	if app.Status.Sync.Status != "" {
		return string(app.Status.Sync.Status)
	}

	// Default status if nothing else is available
	return "Unknown"
}

// getAppHealth returns the health status of the ArgoCD application
func getAppHealth(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}

	// HealthStatus is a struct, we should check if it's empty instead
	if app.Status.Health.Status == "" {
		return "Unknown"
	}
	return string(app.Status.Health.Status)
}

// getAppSyncStatus returns the sync status of the ArgoCD application
func getAppSyncStatus(app *argocdv1alpha1.Application) string {
	if app == nil {
		return "Unknown"
	}
	return string(app.Status.Sync.Status)
}

type projectGitOpsCache struct {
	mu    sync.RWMutex
	cache map[string]string // key: project name, value: gitops repo url
}

var (
	defaultProjectGitOpsCache *projectGitOpsCache
	projectGitOpsCacheOnce    sync.Once
)

func getProjectGitOpsCache() *projectGitOpsCache {
	projectGitOpsCacheOnce.Do(func() {
		log.G().Debug("Initializing project gitops repo cache")
		defaultProjectGitOpsCache = &projectGitOpsCache{
			cache: make(map[string]string),
		}
	})
	return defaultProjectGitOpsCache
}

func (c *projectGitOpsCache) Get(project string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if url, exists := c.cache[project]; exists {
		log.G().WithFields(log.Fields{
			"project": project,
			"url":     url,
		}).Debug("project gitops repo cache hit")
		return url
	}

	// Try to refresh cache on miss
	c.mu.RUnlock()
	if err := refreshProjectGitOpsCache(); err != nil {
		log.G().WithError(err).Error("failed to refresh project gitops cache")
	}
	c.mu.RLock()

	// Check again after refresh
	if url, exists := c.cache[project]; exists {
		log.G().WithFields(log.Fields{
			"project": project,
			"url":     url,
		}).Debug("project gitops repo cache hit after refresh")
		return url
	}

	defaultURL := viper.GetString("application_repo.remote_url")
	log.G().WithFields(log.Fields{
		"project":     project,
		"default_url": defaultURL,
	}).Debug("project gitops repo cache miss, using default")
	return defaultURL
}

func (c *projectGitOpsCache) Set(project, url string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.G().WithFields(log.Fields{
		"project": project,
		"url":     url,
	}).Debug("Setting project gitops repo cache")
	c.cache[project] = url
}

// getGitOpsRepo returns the gitops repo url for the given project
func getGitOpsRepo(project string) string {
	return getProjectGitOpsCache().Get(project)
}

// BuildProjectGitOpsCache builds the cache from project list
func BuildProjectGitOpsCache(tenants []types.TenantInfo) {
	cache := getProjectGitOpsCache()
	for _, tenant := range tenants {
		if tenant.GitOpsRepo != "" {
			cache.Set(tenant.Name, tenant.GitOpsRepo)
			log.G().WithFields(log.Fields{
				"project":   tenant.Name,
				"gitopsURL": tenant.GitOpsRepo,
			}).Debug("cached project gitops repo")
		}
	}
}

// refreshProjectGitOpsCache refreshes the cache from repository
func refreshProjectGitOpsCache() error {
	tenants, err := repowriter.MetaRepo().RunProjectList(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	BuildProjectGitOpsCache(tenants)
	return nil
}

func InitProjectGitOpsCache(tenants []types.TenantInfo) {
	cache := getProjectGitOpsCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// rebuild cache
	for _, tenant := range tenants {
		if tenant.GitOpsRepo != "" {
			cache.cache[tenant.Name] = tenant.GitOpsRepo
			log.G().WithFields(log.Fields{
				"project":   tenant.Name,
				"gitopsURL": tenant.GitOpsRepo,
			}).Debug("cached project gitops repo")
		}
	}

	log.G().WithField("count", len(cache.cache)).Info("Project gitops cache initialized")
}
