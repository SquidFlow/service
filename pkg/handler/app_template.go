package handler

import (
	"github.com/google/uuid"

	apptempv1alpha1 "github.com/h4-poc/argocd-addon/api/v1alpha1"
	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/git"
)

type ActionType string

var (
	ActionTypeCreate ActionType = "create"
	ActionTypeDelete ActionType = "delete"
	ActionTypeUpdate ActionType = "update"
)

type ResourceName string

var (
	ResourceNameApp         ResourceName = "app"
	ResourceNameProject     ResourceName = "project"
	ResourceNameAppTemplate ResourceName = "apptemplate"
)

// ApplicationTemplate represents a template for deploying applications
type ApplicationTemplate struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Path         string                  `json:"path"`
	Validated    bool                    `json:"validated"`
	Owner        string                  `json:"owner"`
	Environments []string                `json:"environments"`
	LastApplied  string                  `json:"lastApplied"`
	AppType      ApplicationTemplateType `json:"appType"`
	Source       ApplicationSource       `json:"source"`
	Resources    ApplicationResources    `json:"resources"`
	Events       []ApplicationEvent      `json:"events"`
	CreatedAt    string                  `json:"createdAt"`
	UpdatedAt    string                  `json:"updatedAt"`
}

type ApplicationSource struct {
	Type           string `json:"type"`
	URL            string `json:"url"`
	TargetRevision string `json:"targetRevision"`
}

type ApplicationResources struct {
	Deployments               int                       `json:"deployments"`
	Services                  int                       `json:"services"`
	Configmaps                int                       `json:"configmaps"`
	Secrets                   int                       `json:"secrets"`
	Ingresses                 int                       `json:"ingresses"`
	ServiceAccounts           int                       `json:"serviceAccounts"`
	Roles                     int                       `json:"roles"`
	RoleBindings              int                       `json:"roleBindings"`
	NetworkPolicies           int                       `json:"networkPolicies"`
	PersistentVolumeClaims    int                       `json:"persistentVolumeClaims"`
	HorizontalPodAutoscalers  int                       `json:"horizontalPodAutoscalers"`
	CustomResourceDefinitions CustomResourceDefinitions `json:"customResourceDefinitions"`
}

type CustomResourceDefinitions struct {
	ExternalSecrets     int `json:"externalSecrets"`
	Certificates        int `json:"certificates"`
	IngressRoutes       int `json:"ingressRoutes"`
	PrometheusRules     int `json:"prometheusRules"`
	ServiceMeshPolicies int `json:"serviceMeshPolicies"`
	VirtualServices     int `json:"virtualServices"`
}

type ApplicationEvent struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type ApplicationTemplateType string

var (
	ApplicationTemplateTypeHelm          ApplicationTemplateType = "helm"
	ApplicationTemplateTypeKustomize     ApplicationTemplateType = "kustomize"
	ApplicationTemplateTypeHelmKustomize ApplicationTemplateType = "helm-kustomize"
)

type (
	AppTemplateCreateOptions struct {
		CloneOpts       *git.CloneOptions
		ProjectName     string
		DestKubeServer  string
		DestKubeContext string
		DryRun          bool
		AddCmd          argocd.AddClusterCmd
		Labels          map[string]string
		Annotations     map[string]string
	}

	AppTemplateListOptions struct {
		CloneOpts *git.CloneOptions
	}
)

func getNewId() string {
	return uuid.New().String()
}

func getAppTempType(temp apptempv1alpha1.ApplicationTemplate) ApplicationTemplateType {
	var enableHelm, enableKustomize bool
	if temp.Spec.Helm != nil {
		enableHelm = true
	}
	if temp.Spec.Kustomize != nil {
		enableKustomize = true
	}

	// only define 2 types: helm and kustomize
	if enableHelm && enableKustomize {
		return ApplicationTemplateTypeHelmKustomize
	}
	if enableHelm {
		return ApplicationTemplateTypeHelm
	}
	if enableKustomize {
		return ApplicationTemplateTypeKustomize
	}
	return "unknown"
}
