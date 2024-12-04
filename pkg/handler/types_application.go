package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/google/uuid"
	"k8s.io/client-go/kubernetes"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/util"
)

// ActionType is the type of action
type ActionType string

const (
	ActionTypeCreate ActionType = "create"
	ActionTypeUpdate ActionType = "update"
	ActionTypeDelete ActionType = "delete"
)

// ResourceName is the name of the resource
type ResourceName string

var (
	ResourceNameApp         ResourceName = "app"
	ResourceNameProject     ResourceName = "project"
	ResourceNameAppTemplate ResourceName = "apptemplate"
)

// validate application source request
type (
	ValidateAppSourceRequest struct {
		Repo          string `json:"repo" binding:"required"`
		TargetVersion string `json:"target_version"`
		Path          string `json:"path" binding:"required"`
	}

	ValidateAppSourceResponse struct {
		Success      bool     `json:"success"`
		Message      string   `json:"message"`
		Type         string   `json:"type"`
		SuiteableEnv []string `json:"suiteable_env"`
	}
)

// create application request and response structs
type (
	// ApplicationSource represents the source information of an application
	ApplicationSource struct {
		Repo                 string                `json:"repo" binding:"required"`
		TargetRevision       string                `json:"target_revision"`
		Path                 string                `json:"path" binding:"required"`
		Submodules           bool                  `json:"submodules,omitempty"`
		ApplicationSpecifier *ApplicationSpecifier `json:"application_specifier,omitempty"`
	}

	// ApplicationSpecifier contains application-specific configuration
	ApplicationSpecifier struct {
		// HelmManifestPath specifies the path to Helm chart manifests
		// Required for Helm applications, must point to the directory containing Chart.yaml
		HelmManifestPath string `json:"helm_manifest_path,omitempty"`
	}

	// ApplicationCreateRequest represents the request body for creating an application
	ApplicationCreateRequest struct {
		// Source information of the application
		ApplicationSource ApplicationSource `json:"application_source" binding:"required"`

		// Application instantiation details
		ApplicationInstantiation ApplicationInstantiation `json:"application_instantiation" binding:"required"`

		// where to deploy the application
		ApplicationTarget []ApplicationTarget `json:"application_target" binding:"required"`

		// Whether this is a dry run
		IsDryRun bool `json:"is_dryrun"`
	}

	// ApplicationInstantiation represents the instantiation details of an application
	ApplicationInstantiation struct {
		ApplicationName string          `json:"application_name" binding:"required"`
		TenantName      string          `json:"tenant_name" binding:"required"`
		AppCode         string          `json:"appcode" binding:"required"`
		Description     string          `json:"description,omitempty"`
		Ingress         []IngressConfig `json:"ingress,omitempty"`
		Security        SecurityConfig  `json:"security,omitempty"`
	}

	// ApplicationTarget represents the target information of an application
	ApplicationTarget struct {
		Cluster   string `json:"cluster" binding:"required"`
		Namespace string `json:"namespace" binding:"required"`
	}

	// IngressConfig represents ingress configuration
	IngressConfig struct {
		Host string    `json:"host,omitempty"`
		TLS  TLSConfig `json:"tls,omitempty"`
	}

	// TLSConfig represents TLS configuration for ingress
	TLSConfig struct {
		Enabled    bool   `json:"enabled,omitempty"`
		SecretName string `json:"secret_name,omitempty"`
	}

	// SecurityConfig represents security configuration
	SecurityConfig struct {
		ExternalSecret ExternalSecretConfig `json:"external_secret,omitempty"`
	}

	// ExternalSecretConfig represents external secret configuration
	ExternalSecretConfig struct {
		SecretStoreRef SecretStoreRefConfig `json:"secret_store_ref"`
	}

	// SecretStoreRefConfig represents secret store reference configuration
	SecretStoreRefConfig struct {
		ID string `json:"id"`
	}

	// AppCreateOptions represents options for creating an application
	AppCreateOptions struct {
		CloneOpts       *git.CloneOptions
		AppsCloneOpts   *git.CloneOptions
		ProjectName     string
		KubeContextName string
		createOpts      *application.CreateOptions
		KubeFactory     kube.Factory
		Timeout         time.Duration
		Labels          map[string]string
		Annotations     map[string]string
		Include         string
		Exclude         string
	}

	// ApplicationCreateResponse represents the response body for creating an application
	ApplicationCreateResponse struct {
		Success     bool        `json:"success"`
		Message     string      `json:"message"`
		Application Application `json:"application"`
	}
)

// list application response struct
type (
	// single application response struct
	Application struct {
		ApplicationSource        ApplicationSource        `json:"application_source" binding:"required"`
		ApplicationInstantiation ApplicationInstantiation `json:"application_instantiation" binding:"required"`
		ApplicationTarget        []ApplicationTarget      `json:"application_target" binding:"required"`
		DryrunResult             ApplicationDryRunResult  `json:"dryrun_result,omitempty"`
		ApplicationRuntime       ApplicationRuntime       `json:"application_runtime,omitempty"`
	}

	// ApplicationDryRunResult represents the dry run result of an application
	// manifest and argocd file
	ApplicationDryRunResult struct {
		Success      bool                   `json:"success"`
		Message      string                 `json:"message"`
		Total        int                    `json:"total"`
		Environments []ApplicationDryRunEnv `json:"environments"`
	}

	// ApplicationDryRunEnv represents the dry run result for each environment
	ApplicationDryRunEnv struct {
		Environment string `json:"environment"`
		IsValid     bool   `json:"is_valid"`
		Manifest    string `json:"manifest,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	// ApplicationRuntime represents the runtime information of an application
	ApplicationRuntime struct {
		Status          string              `json:"status"`
		Health          string              `json:"health"`
		SyncStatus      string              `json:"sync_status"`
		GitInfo         []GitInfo           `json:"git_info"`
		ResourceMetrics ResourceMetricsInfo `json:"resource_metrics"`
	}

	ResourceMetricsInfo struct {
		PodCount    int    `json:"pod_count"`
		SecretCount int    `json:"secret_count"`
		CPU         string `json:"cpu"`
		Memory      string `json:"memory"`
	}

	// ApplicationListResponse represents the response body for listing applications
	ApplicationListResponse struct {
		Total        int64         `json:"total"`
		Success      bool          `json:"success"`
		Message      string        `json:"message"`
		Applications []Application `json:"applications"`
	}
)

// ApplicationUpdateRequest represents the request body for updating an application
type (
	ApplicationUpdateRequest struct {
		ApplicationSource        ApplicationSource        `json:"application_source" binding:"omitempty"`
		ApplicationInstantiation ApplicationInstantiation `json:"application_instantiation" binding:"omitempty"`
		ApplicationTarget        []ApplicationTarget      `json:"application_target" binding:"omitempty"`
	}

	ApplicationUpdateResponse struct {
		Success     bool        `json:"success"`
		Message     string      `json:"message"`
		Application Application `json:"application"`
	}
)

type (
	SyncApplicationRequest struct {
		Applications []string `json:"applications" binding:"required,min=1"`
	}

	SyncApplicationResponse struct {
		Results []SyncApplicationResult `json:"results"`
	}

	SyncApplicationResult struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}
)

type GitInfo struct {
	Creator           string
	LastUpdater       string
	LastCommitID      string
	LastCommitMessage string
}

// TODO: Implement this function later
func getGitInfo(repofs billy.Filesystem, appPath string) (*GitInfo, error) {
	return &GitInfo{
		Creator:           "Unknown",
		LastUpdater:       "Unknown",
		LastCommitID:      "Unknown",
		LastCommitMessage: "Unknown",
	}, nil
}

// TODO: Implement this function later
func getResourceMetrics(ctx context.Context, kubeClient kubernetes.Interface, namespace string) (*ResourceMetricsInfo, error) {
	return &ResourceMetricsInfo{
		PodCount:    5,
		SecretCount: 12,
		CPU:         "0.25",
		Memory:      "200Mi",
	}, nil
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

// getNewId returns a new id for the resource
func getNewId() string {
	return uuid.New().String()
}

// ValidationRequest represents the request structure for template validation
type ValidationRequest struct {
	Source         ApplicationSource `json:"source" binding:"required"`
	Path           string            `json:"path" binding:"required"`
	TargetRevision string            `json:"targetRevision" binding:"required"`
}

// ValidationResult represents the validation result for each environment
type ValidationResult struct {
	Environment string   `json:"environment"` // the repo support multiple environments
	IsValid     bool     `json:"isValid"`
	Message     []string `json:"message,omitempty"`
}

// ValidationResponse represents the response structure for template validation
type ValidationResponse struct {
	Success bool               `json:"success"`
	Error   string             `json:"error,omitempty"`
	Results []ValidationResult `json:"results"`
}

var setAppOptsDefaults = func(ctx context.Context, repofs fs.FS, opts *AppCreateOptions) error {
	var err error

	if opts.createOpts.DestServer == store.Default.DestServer || opts.createOpts.DestServer == "" {
		opts.createOpts.DestServer, err = getProjectDestServer(repofs, opts.ProjectName)
		if err != nil {
			return err
		}
	}

	if opts.createOpts.DestNamespace == "" {
		opts.createOpts.DestNamespace = "default"
	}

	if opts.createOpts.Labels == nil {
		opts.createOpts.Labels = opts.createOpts.Labels
	}

	if opts.createOpts.Annotations == nil {
		opts.createOpts.Annotations = opts.createOpts.Annotations
	}

	if opts.createOpts.AppType != "" {
		return nil
	}

	var fsys fs.FS
	if _, err := os.Stat(opts.createOpts.AppSpecifier); err == nil {
		// local directory
		fsys = fs.Create(osfs.New(opts.createOpts.AppSpecifier))
	} else {
		host, orgRepo, p, _, _, suffix, _ := util.ParseGitUrl(opts.createOpts.AppSpecifier)
		url := host + orgRepo + suffix
		log.G().Infof("cloning repo: '%s', to infer app type from path '%s'", url, p)
		cloneOpts := &git.CloneOptions{
			Repo:     opts.createOpts.AppSpecifier,
			Auth:     opts.CloneOpts.Auth,
			Provider: opts.CloneOpts.Provider,
			FS:       fs.Create(memfs.New()),
		}
		cloneOpts.Parse()
		_, fsys, err = getRepo(ctx, cloneOpts)
		if err != nil {
			return err
		}
	}

	opts.createOpts.AppType = application.InferAppType(fsys)
	log.G().Infof("inferred application type: %s", opts.createOpts.AppType)

	return nil
}

var parseApp = func(appOpts *application.CreateOptions, projectName, repoURL, targetRevision, repoRoot string) (application.Application, error) {
	return appOpts.Parse(projectName, repoURL, targetRevision, repoRoot)
}

func getProjectDestServer(repofs fs.FS, projectName string) (string, error) {
	path := repofs.Join(store.Default.ProjectsDir, projectName+".yaml")
	p := &argocdv1alpha1.AppProject{}
	if err := repofs.ReadYamls(path, p); err != nil {
		return "", fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return p.Annotations[store.Default.DestServerAnnotation], nil
}

func genCommitMsg(action ActionType, targetResource ResourceName, appName, projectName string, repofs fs.FS) string {
	commitMsg := fmt.Sprintf("%s %s '%s' on project '%s'", action, targetResource, appName, projectName)
	if repofs.Root() != "" {
		commitMsg += fmt.Sprintf(" installation-path: '%s'", repofs.Root())
	}

	return commitMsg
}

func getConfigFileFromPath(repofs fs.FS, appPath string) (*application.Config, error) {
	path := repofs.Join(appPath, "config.json")
	b, err := repofs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s'", path)
	}

	conf := application.Config{}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file '%s'", path)
	}

	return &conf, nil
}

var getInstallationNamespace = func(repofs fs.FS) (string, error) {
	path := repofs.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName+".yaml")
	a := &argocdv1alpha1.Application{}
	if err := repofs.ReadYamls(path, a); err != nil {
		return "", fmt.Errorf("failed to unmarshal namespace: %w", err)
	}

	return a.Spec.Destination.Namespace, nil
}

func waitAppSynced(ctx context.Context, f kube.Factory, timeout time.Duration, appName, namespace, revision string, waitForCreation bool) error {
	return f.Wait(ctx, &kube.WaitOptions{
		Interval: store.Default.WaitInterval,
		Timeout:  timeout,
		Resources: []kube.Resource{
			{
				Name:      appName,
				Namespace: namespace,
				WaitFunc:  argocd.GetAppSyncWaitFunc(revision, waitForCreation),
			},
		},
	})
}
