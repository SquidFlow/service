package types

import (
	"time"

	argocdv1alpha1client "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"k8s.io/client-go/kubernetes"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
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
	ApplicationSourceRequest struct {
		Repo                 string                `json:"repo" binding:"required"`
		TargetRevision       string                `json:"target_revision,omitempty"`
		Path                 string                `json:"path,omitempty"`
		Submodules           bool                  `json:"submodules,omitempty"`
		ApplicationSpecifier *ApplicationSpecifier `json:"application_specifier,omitempty"`
	}

	// ApplicationSpecifier contains application-specific configuration
	ApplicationSpecifier struct {
		// HelmManifestPath specifies the path to Helm chart manifests
		// Required for Helm applications, must point to the directory containing Chart.yaml
		HelmManifestPath string `json:"helm_manifest_path,omitempty"`
	}

	ValidateAppSourceResponse struct {
		Success      bool                       `json:"success"`
		Message      string                     `json:"message"`
		Type         string                     `json:"type"`
		SuiteableEnv []AppSourceWithEnvironment `json:"suiteable_env"`
	}

	AppSourceWithEnvironment struct {
		Environments string `json:"environments"`
		Valid        bool   `json:"valid"`
		Error        string `json:"error,omitempty"`
	}
)

// create application request and response structs
type (
	// ApplicationCreateRequest represents the request body for creating an application
	ApplicationCreateRequest struct {
		// Source information of the application
		ApplicationSource ApplicationSourceRequest `json:"application_source" binding:"required"`

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
		AppOpts         *application.CreateOptions
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
		ApplicationSource        ApplicationSourceRequest `json:"application_source" binding:"required"`
		ApplicationInstantiation ApplicationInstantiation `json:"application_instantiation" binding:"required"`
		ApplicationTarget        []ApplicationTarget      `json:"application_target" binding:"required"`
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
		ArgocdFile  string `json:"argocd_file,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	// ApplicationRuntime represents the runtime information of an application
	ApplicationRuntime struct {
		Status          string              `json:"status"`
		Health          string              `json:"health"`
		SyncStatus      string              `json:"sync_status"`
		GitInfo         []GitInfo           `json:"git_info"`
		ResourceMetrics ResourceMetricsInfo `json:"resource_metrics"`
		ArgoCDUrl       string              `json:"argocd_url"`
		CreatedAt       time.Time           `json:"created_at"`
		CreatedBy       string              `json:"created_by"`
		LastUpdatedAt   time.Time           `json:"last_updated_at"`
		LastUpdatedBy   string              `json:"last_updated_by"`
	}

	ResourceMetricsInfo struct {
		PodCount    int    `json:"pod_count"`
		SecretCount int    `json:"secret_count"`
		CPU         string `json:"cpu"`
		Memory      string `json:"memory"`
	}

	// ApplicationListResponse represents the response body for listing applications
	ApplicationListResponse struct {
		Total   int64         `json:"total"`
		Success bool          `json:"success"`
		Message string        `json:"message"`
		Error   string        `json:"error,omitempty"`
		Items   []Application `json:"items"`
	}
)

// ApplicationUpdateRequest represents the request body for updating an application
type (
	ApplicationUpdateRequest struct {
		ApplicationSource        ApplicationSourceRequest `json:"application_source" binding:"omitempty"`
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

// ValidationRequest represents the request structure for template validation
type ValidationRequest struct {
	Source         ApplicationSourceRequest `json:"source" binding:"required"`
	Path           string                   `json:"path" binding:"required"`
	TargetRevision string                   `json:"targetRevision" binding:"required"`
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

// Delete
type (
	AppDeleteOptions struct {
		CloneOpts   *git.CloneOptions
		ProjectName string
		AppName     string
		Global      bool
	}
)

// list
type (
	AppListOptions struct {
		CloneOpts    *git.CloneOptions
		ProjectName  string
		KubeClient   kubernetes.Interface
		ArgoCDClient *argocdv1alpha1client.ArgoprojV1alpha1Client
	}
)

// update
type (
	UpdateOptions struct {
		CloneOpts   *git.CloneOptions
		ProjectName string
		AppName     string
		Username    string
		UpdateReq   *ApplicationUpdateRequest
		KubeFactory kube.Factory
		Annotations map[string]string
	}
)

// get
type (
	AppGetOptions struct {
		CloneOpts   *git.CloneOptions
		ProjectName string
		AppName     string
	}
)
