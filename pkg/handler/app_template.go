package handler

// ApplicationTemplate represents a template for deploying applications
type ApplicationTemplate struct {
	ID           int                  `json:"id"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Path         string               `json:"path"`
	Validated    bool                 `json:"validated"`
	Owner        string               `json:"owner"`
	Environments []string             `json:"environments"`
	LastApplied  string               `json:"lastApplied"`
	AppType      string               `json:"appType"`
	Source       ApplicationSource    `json:"source"`
	Resources    ApplicationResources `json:"resources"`
	Events       []ApplicationEvent   `json:"events"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

type ApplicationSource struct {
	Type   string `json:"type"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
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
