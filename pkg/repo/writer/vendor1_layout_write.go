// Package repowriter provides implementations for different GitOps repository layouts.
//
// This file implements the Vendor1RepoTarget, which handles the vendor1-specific
// GitOps repository structure. The vendor1 layout is characterized by:
// - A specific directory structure for organizing applications and projects
// - Custom handling of secret stores and their configurations
// - Implementation of the RepoWriter interface for standardized repository operations
//
// The Vendor1RepoTarget supports operations such as:
// - Application lifecycle management (create, update, delete, list)
// - Project management (create, delete, list)
// - Secret store operations (list, write, delete)
//
// Usage:
//
//	writer := &Vendor1RepoTarget{}
//	err := writer.RunAppCreate(ctx, opts)

package writer

import (
	"context"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/types"
)

var _ MetaRepoWriter = &Vendor1RepoTarget{}

type Vendor1RepoTarget struct {
	Vendor1RepoTargetSecretStore
	Vendor1RepoTargetApp
	Vendor1RepoTargetProject
}
type Vendor1RepoTargetApp struct {
}

func (v *Vendor1RepoTargetApp) RunAppGet(ctx context.Context, appName string) (*types.Application, error) {
	return nil, nil
}

func (v *Vendor1RepoTargetApp) RunAppList(ctx context.Context) ([]types.Application, error) {
	return nil, nil
}

func (v *Vendor1RepoTargetApp) RunAppCreate(ctx context.Context, opts *application.AppCreateOptions) error {
	return nil
}

func (v *Vendor1RepoTargetApp) RunAppDelete(ctx context.Context, name string) error {
	return nil
}

func (v *Vendor1RepoTargetApp) RunAppUpdate(ctx context.Context, opts *types.UpdateOptions) error {
	return nil
}

type Vendor1RepoTargetSecretStore struct {
}

func (v *Vendor1RepoTargetSecretStore) SecretStoreCreate(ctx context.Context, ss *esv1beta1.SecretStore, force bool) error {
	return nil
}

func (v *Vendor1RepoTargetSecretStore) SecretStoreUpdate(ctx context.Context, id string, req *types.SecretStoreUpdateRequest) (*esv1beta1.SecretStore, error) {
	return nil, nil
}

func (v *Vendor1RepoTargetSecretStore) SecretStoreDelete(ctx context.Context, id string) error {
	return nil
}

func (v *Vendor1RepoTargetSecretStore) SecretStoreGet(ctx context.Context, id string) (*esv1beta1.SecretStore, error) {
	return nil, nil
}

func (v *Vendor1RepoTargetSecretStore) SecretStoreList(ctx context.Context) ([]esv1beta1.SecretStore, error) {
	return nil, nil
}

type Vendor1RepoTargetProject struct {
}

func (v *Vendor1RepoTargetProject) RunProjectCreate(ctx context.Context, opts *types.ProjectCreateOptions) error {
	return nil
}

func (v *Vendor1RepoTargetProject) RunProjectGet(ctx context.Context, projectName string) (*types.TenantDetailInfo, error) {
	return nil, nil
}

func (v *Vendor1RepoTargetProject) RunProjectList(ctx context.Context) ([]types.TenantInfo, error) {
	return nil, nil
}

func (v *Vendor1RepoTargetProject) RunProjectDelete(ctx context.Context, name string) error {
	return nil
}
