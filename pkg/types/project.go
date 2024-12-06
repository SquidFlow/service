package types

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"io"

	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/git"
)

type (
	ProjectCreateOptions struct {
		CloneOpts       *git.CloneOptions
		ProjectName     string
		DestKubeServer  string
		DestKubeContext string
		DryRun          bool
		AddCmd          argocd.AddClusterCmd
		Labels          map[string]string
		Annotations     map[string]string
	}

	ProjectDeleteOptions struct {
		CloneOpts   *git.CloneOptions
		ProjectName string
	}

	ProjectListOptions struct {
		CloneOpts *git.CloneOptions
		Out       io.Writer
	}

	GenerateProjectOptions struct {
		Name               string
		Namespace          string
		DefaultDestServer  string
		DefaultDestContext string
		RepoURL            string
		Revision           string
		InstallationPath   string
		Labels             map[string]string
		Annotations        map[string]string
	}
)

type CreateAppSetOptions struct {
	name                        string
	namespace                   string
	appName                     string
	appNamespace                string
	appProject                  string
	repoURL                     string
	revision                    string
	srcPath                     string
	destServer                  string
	destNamespace               string
	prune                       bool
	preserveResourcesOnDeletion bool
	appLabels                   map[string]string
	appAnnotations              map[string]string
	generators                  []argocdv1alpha1.ApplicationSetGenerator
}

type TenantInfo struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	DefaultCluster string `json:"default_cluster"`
}

type TenantDetailInfo struct {
	Name                       string            `json:"name"`
	Namespace                  string            `json:"namespace"`
	Description                string            `json:"description,omitempty"`
	DefaultCluster             string            `json:"default_cluster"`
	SourceRepos                []string          `json:"source_repos,omitempty"`
	Destinations               []ProjectDest     `json:"destinations,omitempty"`
	ClusterResourceWhitelist   []ProjectResource `json:"cluster_resource_whitelist,omitempty"`
	NamespaceResourceWhitelist []ProjectResource `json:"namespace_resource_whitelist,omitempty"`
	CreatedBy                  string            `json:"created_by"`
	CreatedAt                  string            `json:"created_at,omitempty"`
}

type ProjectDest struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

type ProjectResource struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
}

type ProjectCreateRequest struct {
	ProjectName string            `json:"project-name" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}
