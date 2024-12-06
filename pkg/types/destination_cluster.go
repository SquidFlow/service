package types

// ClusterInfo represents a Kubernetes cluster's information
type ClusterInfo struct {
	Name              string         `json:"name"`
	Environment       string         `json:"environment"`
	Status            string         `json:"status"`
	Provider          string         `json:"provider"`
	Version           ClusterVersion `json:"version"`
	NodeCount         int            `json:"nodeCount"`
	Region            string         `json:"region"`
	ResourceQuota     ResourceQuota  `json:"resourceQuota"`
	Health            HealthStatus   `json:"health"`
	Nodes             NodeStatus     `json:"nodes"`
	NetworkPolicy     bool           `json:"networkPolicy"`
	IngressController string         `json:"ingressController"`
	LastUpdated       string         `json:"lastUpdated"`
	ConsoleURL        string         `json:"consoleUrl,omitempty"`
	Monitoring        Monitoring     `json:"monitoring"`
	Builtin           bool           `json:"builtin,omitempty"`
}

type ClusterVersion struct {
	Kubernetes string `json:"kubernetes"`
	Platform   string `json:"platform"`
}

type Monitoring struct {
	Prometheus   bool        `json:"prometheus"`
	Grafana      bool        `json:"grafana"`
	Alertmanager bool        `json:"alertmanager"`
	URLs         MonitorURLs `json:"urls,omitempty"`
}

type MonitorURLs struct {
	Prometheus   string `json:"prometheus,omitempty"`
	Grafana      string `json:"grafana,omitempty"`
	Alertmanager string `json:"alertmanager,omitempty"`
}

// CreateClusterRequest represents the request body for cluster creation
type CreateClusterRequest struct {
	Name       string            `json:"name" binding:"required"`
	Env        string            `json:"env" binding:"required,oneof=DEV SIT UAT PRD"`
	KubeConfig string            `json:"kubeconfig" binding:"required"` // with base64 encoding
	Labels     map[string]string `json:"labels,omitempty"`              // custom labels
}

// TLSClientConfig represents the structure of the config data in the secret
type TLSClientConfig struct {
	TLSClientConfig struct {
		Insecure bool   `json:"insecure"`
		CertData string `json:"certData"`
		KeyData  string `json:"keyData"`
		CAData   string `json:"caData"`
	} `json:"tlsClientConfig"`
}

type ClusterResponse struct {
	Name              string            `json:"name"`
	Environment       string            `json:"environment"`
	Status            []ComponentStatus `json:"componentStatus"`
	Provider          string            `json:"provider"`
	Version           VersionInfo       `json:"version"`
	NodeCount         int               `json:"nodeCount"`
	Region            string            `json:"region"`
	ResourceQuota     ResourceQuota     `json:"resourceQuota"`
	Health            HealthStatus      `json:"health"`
	Nodes             NodeStatus        `json:"nodes"`
	NetworkPolicy     bool              `json:"networkPolicy"`
	IngressController string            `json:"ingressController"`
	LastUpdated       string            `json:"lastUpdated"`
	ConsoleUrl        string            `json:"consoleUrl,omitempty"`
	Monitoring        MonitoringInfo    `json:"monitoring"`
	Builtin           bool              `json:"builtin,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
}

type VersionInfo struct {
	Kubernetes string `json:"kubernetes"`
	Platform   string `json:"platform"`
}

type ResourceQuota struct {
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
	Storage   string `json:"storage"`
	PVCs      string `json:"pvcs"`
	NodePorts string `json:"nodeports"`
}

type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type NodeStatus struct {
	Ready int `json:"ready"`
	Total int `json:"total"`
}

type MonitoringInfo struct {
	Prometheus   bool            `json:"prometheus"`
	Grafana      bool            `json:"grafana"`
	AlertManager bool            `json:"alertmanager"`
	URLs         *MonitoringURLs `json:"urls,omitempty"`
}

type MonitoringURLs struct {
	Prometheus   string `json:"prometheus,omitempty"`
	Grafana      string `json:"grafana,omitempty"`
	AlertManager string `json:"alertmanager,omitempty"`
}

type ComponentStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ClusterListResponse struct {
	Total   int               `json:"total"`
	Message string            `json:"message"`
	Items   []ClusterResponse `json:"items"`
}

// UpdateClusterRequest represents the request body for cluster update
type UpdateClusterRequest struct {
	Env        string            `json:"env" binding:"required,oneof=DEV SIT UAT PRD"`
	KubeConfig string            `json:"kubeconfig,omitempty"` // base64 encoded, optional
	Labels     map[string]string `json:"labels,omitempty"`     // custom labels
}
