package repotarget

import (
	"context"

	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/types"
)

// RepoWriter defines how to interact with a GitOps repository
type RepoWriter interface {
	ApplicationWriter
	ProjectWriter
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
