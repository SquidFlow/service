package repowriter

import (
	"context"
	"fmt"
	"sync"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

var (
	instance RepoWriter
	once     sync.Once
	initErr  error
)

// this is native repo layout
// tree -L 1
// .
// ├── apps
// ├── bootstrap
// └── projects
// this is vendor1 repo layout
// tree -L 1
// .
// ├── overlays
// └── manifest
// InitRepoWriter initializes the RepoWriter based on the repository layout
// This should be called once during application startup
func InitRepoWriter(fs fs.FS) error {
	once.Do(func() {
		// Check native layout paths
		appsExists, appsErr := fs.Exists("apps/")
		bootstrapExists, bootstrapErr := fs.Exists("bootstrap/")
		projectsExists, projectsErr := fs.Exists("projects/")
		// Check vendor1 layout paths
		overlaysExists, overlaysErr := fs.Exists("overlays/")
		manifestExists, manifestErr := fs.Exists("manifest/")

		// Handle any filesystem errors first
		if appsErr != nil || bootstrapErr != nil || projectsErr != nil ||
			overlaysErr != nil || manifestErr != nil {
			initErr = fmt.Errorf("failed to check repository layout: apps(%v), bootstrap(%v), projects(%v), overlays(%v), manifest(%v)",
				appsErr, bootstrapErr, projectsErr, overlaysErr, manifestErr)
			return
		}

		switch {
		case appsExists && bootstrapExists && projectsExists:
			instance = &NativeRepoTarget{}
			log.G().Info("using native repo layout")
		case overlaysExists && manifestExists:
			instance = &Vendor1RepoTarget{}
			log.G().Info("using vendor1 repo layout")
		default:
			initErr = fmt.Errorf("not supported repo layout")
		}
	})
	return initErr
}

// GetRepoWriter returns the initialized RepoWriter instance
func GetRepoWriter() RepoWriter {
	return instance
}

// RepoWriter defines how to interact with a GitOps repository
type RepoWriter interface {
	ApplicationWriter
	ProjectWriter
	SecretStoreWriter
}

type ApplicationWriter interface {
	// RunAppCreate creates an application
	RunAppCreate(ctx context.Context, opts *types.AppCreateOptions) error
	// RunAppDelete deletes an application
	RunAppDelete(ctx context.Context, opts *types.AppDeleteOptions) error
	// RunAppUpdate updates an application
	RunAppUpdate(ctx context.Context, opts *types.UpdateOptions) error
	// RunAppGet gets a single application
	RunAppGet(ctx context.Context, opts *types.AppListOptions, appName string) (*types.Application, error)
	// RunAppList lists all applications
	RunAppList(ctx context.Context, opts *types.AppListOptions) (*types.ApplicationListResponse, error)
}

type ProjectWriter interface {
	// RunProjectCreate Project methods
	RunProjectCreate(ctx context.Context, opts *types.ProjectCreateOptions) error
	// RunProjectGetDetail gets a single project
	RunProjectGetDetail(ctx context.Context, projectName string, opts *git.CloneOptions) (*types.TenantDetailInfo, error)
	// RunProjectList lists all projects
	RunProjectList(ctx context.Context, opts *types.ProjectListOptions) ([]types.TenantInfo, error)
	// RunProjectDelete deletes a project
	RunProjectDelete(ctx context.Context, opts *types.ProjectDeleteOptions) error
}

type SecretStoreWriter interface {
	WriteSecretStore2Repo(ctx context.Context, ss *esv1beta1.SecretStore, cloneOpts *git.CloneOptions, force bool) error
	RunDeleteSecretStore(ctx context.Context, secretStoreID string, opts *types.SecretStoreDeleteOptions) error
	GetSecretStoreFromRepo(ctx context.Context, opts *types.SecretStoreGetOptions) (*types.SecretStoreDetail, error)
	RunListSecretStore(ctx context.Context, opts *types.SecretStoreListOptions) ([]types.SecretStoreDetail, error)
}
