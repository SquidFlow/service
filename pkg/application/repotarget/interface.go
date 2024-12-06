package repotarget

import (
	"context"

	"github.com/squidflow/service/pkg/types"
)

// RepoTarget defines how to interact with a GitOps repository
type RepoTarget interface {
	// Create creates an application
	RunAppCreate(ctx context.Context, opts *types.AppCreateOptions) error
	// Delete deletes an application
	RunAppDelete(ctx context.Context, opts *types.AppDeleteOptions) error
	// Update updates an application
	RunAppUpdate(ctx context.Context, opts *types.UpdateOptions) error
	// Get gets a single application
	RunAppGet(ctx context.Context, opts *types.AppListOptions, appName string) (*types.Application, error)
	// List lists all applications
	RunAppList(ctx context.Context, opts *types.AppListOptions) (*types.ApplicationListResponse, error)
}
