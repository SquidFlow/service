package writer

import (
	"context"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/types"
)

// errorRepoTarget implements the RepoTarget interface, all methods return the same error
type errorRepoWriter struct {
	err error
}

func (e *errorRepoWriter) RunAppCreate(ctx context.Context, opts *application.AppCreateOptions) (*types.ApplicationCreatedResp, error) {
	return nil, e.err
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
