package repowriter

import (
	"context"
	"fmt"
	"sync"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/viper"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

var (
	metarepo    MetaRepoWriter
	tenantRepos sync.Map // key: tenant name, value: TenantRepoWriter
	once        sync.Once
	initErr     error
)

// this is native repo layout
// tree -L 1
// .
// ├── apps
// ├── bootstrap
// └── projects
//
// this is vendor1 repo layout
// tree -L 1
// .
// ├── overlays
// └── manifest
// InitMetaRepoWriter initializes the RepoWriter based on the repository layout
// This should be called once during application startup
func BuildMetaRepoWriter(repofs fs.FS) error {
	once.Do(func() {
		// Check native layout paths
		appsExists, appsErr := repofs.Exists("apps/")
		bootstrapExists, bootstrapErr := repofs.Exists("bootstrap/")
		projectsExists, projectsErr := repofs.Exists("projects/")

		// Check vendor1 layout paths
		overlaysExists, overlaysErr := repofs.Exists("overlays/")
		manifestExists, manifestErr := repofs.Exists("manifest/")

		// Handle any filesystem errors first
		if appsErr != nil || bootstrapErr != nil || projectsErr != nil ||
			overlaysErr != nil || manifestErr != nil {
			initErr = fmt.Errorf("failed to check repository layout: apps(%v), bootstrap(%v), projects(%v), overlays(%v), manifest(%v)",
				appsErr, bootstrapErr, projectsErr, overlaysErr, manifestErr)
			return
		}

		switch {
		case appsExists && bootstrapExists && projectsExists:
			native := &NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo:     viper.GetString("application_repo.remote_url"),
					FS:       fs.Create(memfs.New()),
					Provider: "github",
					Auth: git.Auth{
						Password: viper.GetString("application_repo.access_token"),
					},
					CloneForWrite: true,
				},
				tenantRepoCloneOpts: &git.CloneOptions{
					Repo:     viper.GetString("application_repo.remote_url"),
					FS:       fs.Create(memfs.New()),
					Provider: "github",
					Auth: git.Auth{
						Password: viper.GetString("application_repo.access_token"),
					},
					CloneForWrite: true,
				},
			}
			native.metaRepoCloneOpts.Parse()
			native.tenantRepoCloneOpts.Parse()
			metarepo = native
			log.G().Info("using native repo layout")
		case overlaysExists && manifestExists:
			metarepo = &Vendor1RepoTarget{}
			log.G().Info("using vendor1 repo layout")

		default:
			initErr = fmt.Errorf("not supported repo layout")
		}
	})
	return initErr
}

func buildTenantRepoWriter(tenant types.TenantInfo) TenantRepoWriter {
	log.G().WithFields(log.Fields{
		"tenant": tenant.Name,
		"repo":   tenant.GitOpsRepo,
	}).Debug("building tenant repo writer")

	// if tenant.GitOpsRepo is the same as the application repo, we don't need to clone it
	if tenant.GitOpsRepo == viper.GetString("application_repo.remote_url") {
		log.G().WithFields(log.Fields{
			"tenant": tenant.Name,
			"repo":   tenant.GitOpsRepo,
		}).Debug("skip building tenant repo writer, use meta repo for tenant")
		return metarepo
	}

	tenantRepoCloneOpts := &git.CloneOptions{
		Repo:     tenant.GitOpsRepo,
		FS:       fs.Create(memfs.New()),
		Provider: "github",
		Auth: git.Auth{
			Password: viper.GetString("application_repo.access_token"),
		},
		CloneForWrite: true,
	}
	tenantRepoCloneOpts.Parse()

	_, repofs, err := tenantRepoCloneOpts.GetRepo(context.Background())
	if err != nil {
		log.G().Errorf("failed to get git repo: %v", err)
		return nil
	}

	// Check native layout paths
	appsExists, appsErr := repofs.Exists("apps/")

	// Check vendor1 layout paths
	overlaysExists, overlaysErr := repofs.Exists("overlays/")
	manifestExists, manifestErr := repofs.Exists("manifest/")

	// Handle any filesystem errors first
	if appsErr != nil || overlaysErr != nil || manifestErr != nil {
		initErr = fmt.Errorf("failed to check repository layout: apps(%v), overlays(%v), manifest(%v)",
			appsErr, overlaysErr, manifestErr)
		return nil
	}

	switch {
	case appsExists:
		native := &NativeRepoTarget{
			project: tenant.Name,
			metaRepoCloneOpts: &git.CloneOptions{
				Repo:     viper.GetString("application_repo.remote_url"),
				FS:       fs.Create(memfs.New()),
				Provider: "github",
				Auth: git.Auth{
					Password: viper.GetString("application_repo.access_token"),
				},
				CloneForWrite: true,
			},
			tenantRepoCloneOpts: tenantRepoCloneOpts,
		}
		native.metaRepoCloneOpts.Parse()
		native.tenantRepoCloneOpts.Parse()
		return native
	case overlaysExists && manifestExists:
		return &Vendor1RepoTarget{}
	default:
		initErr = fmt.Errorf("not supported repo layout")
	}

	log.G().WithFields(log.Fields{
		"tenant": tenant.Name,
		"repo":   tenant.GitOpsRepo,
	}).Warn("failed to build tenant repo writer")

	return nil
}

// MetaRepo returns the initialized RepoWriter instance
func MetaRepo() MetaRepoWriter {
	if metarepo == nil {
		log.G().Fatal("meta repo writer is not initialized")
	}
	return metarepo
}

// BuildTenantRepo creates or gets a TenantRepoWriter for the given tenant
func BuildTenantRepo() error {
	tenants, err := metarepo.RunProjectList(context.Background())

	if err != nil {
		log.G().Errorf("failed to list tenants: %v", err)
		return nil
	}

	for _, tenant := range tenants {
		tenantRepoWriter := buildTenantRepoWriter(tenant)
		if tenantRepoWriter == nil {
			log.G().WithFields(log.Fields{
				"tenant": tenant.Name,
				"repo":   tenant.GitOpsRepo,
			}).Warn("invalid tenant repo writer")
			return nil
		}
		log.G().WithField("tenant", tenant.Name).Debug("stored tenant repo writer")
		tenantRepos.Store(tenant.Name, tenantRepoWriter)
	}
	return nil
}

// TenantRepo removes the TenantRepoWriter for the given tenant
func TenantRepo(name string) TenantRepoWriter {
	tenantRepo, ok := tenantRepos.Load(name)
	if !ok {
		// return a special RepoTarget, its all methods will return tenant not found error
		log.G().WithField("tenant", name).Warn("tenant not found, return error repo writer")
		return &errorRepoWriter{err: fmt.Errorf("tenant '%s' not found", name)}
	}
	return tenantRepo.(TenantRepoWriter)
}

// errorRepoTarget implements the RepoTarget interface, all methods return the same error
type errorRepoWriter struct {
	err error
}

func (e *errorRepoWriter) RunAppCreate(ctx context.Context, opts *types.AppCreateOptions) error {
	return e.err
}

func (e *errorRepoWriter) RunAppDelete(ctx context.Context, name string) error {
	return e.err
}

func (e *errorRepoWriter) RunAppUpdate(ctx context.Context, opts *types.UpdateOptions) error {
	return e.err
}

func (e *errorRepoWriter) RunAppGet(ctx context.Context, appName string) (*types.Application, error) {
	return nil, e.err
}

func (e *errorRepoWriter) RunAppList(ctx context.Context) ([]types.Application, error) {
	return nil, e.err
}

func (e *errorRepoWriter) SecretStoreCreate(ctx context.Context, ss *esv1beta1.SecretStore, force bool) error {
	return e.err
}

func (e *errorRepoWriter) SecretStoreUpdate(ctx context.Context, id string, req *types.SecretStoreUpdateRequest) (*esv1beta1.SecretStore, error) {
	return nil, e.err
}

func (e *errorRepoWriter) SecretStoreDelete(ctx context.Context, id string) error {
	return e.err
}

func (e *errorRepoWriter) SecretStoreGet(ctx context.Context, id string) (*esv1beta1.SecretStore, error) {
	return nil, e.err
}

func (e *errorRepoWriter) SecretStoreList(ctx context.Context) ([]esv1beta1.SecretStore, error) {
	return nil, e.err
}

// MetaRepoWriter defines how to interact with a GitOps repository
type MetaRepoWriter interface {
	ApplicationWriter
	ProjectWriter
	SecretStoreWriter
}

// TenantRepoWriter is a repo writer for tenant
type TenantRepoWriter interface {
	ApplicationWriter
	SecretStoreWriter
}

type ApplicationWriter interface {
	RunAppCreate(ctx context.Context, opts *types.AppCreateOptions) error
	RunAppGet(ctx context.Context, name string) (*types.Application, error)
	RunAppDelete(ctx context.Context, name string) error
	RunAppUpdate(ctx context.Context, opts *types.UpdateOptions) error
	RunAppList(ctx context.Context) ([]types.Application, error)
}

// ProjectWriter defines how to interact with a GitOps repository
type ProjectWriter interface {
	RunProjectCreate(ctx context.Context, opts *types.ProjectCreateOptions) error
	RunProjectGet(ctx context.Context, name string) (*types.TenantDetailInfo, error)
	RunProjectList(ctx context.Context) ([]types.TenantInfo, error)
	RunProjectDelete(ctx context.Context, name string) error
}

type SecretStoreWriter interface {
	SecretStoreCreate(ctx context.Context, ss *esv1beta1.SecretStore, override bool) error
	SecretStoreUpdate(ctx context.Context, id string, req *types.SecretStoreUpdateRequest) (*esv1beta1.SecretStore, error)
	SecretStoreDelete(ctx context.Context, id string) error
	SecretStoreGet(ctx context.Context, id string) (*esv1beta1.SecretStore, error)
	SecretStoreList(ctx context.Context) ([]esv1beta1.SecretStore, error)
}
