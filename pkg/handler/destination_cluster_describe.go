package handler

import (
	"context"
	"fmt"

	clientCluster "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/gin-gonic/gin"

	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/log"
)

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

// GetDestinationCluster handles the GET request for a single cluster
func GetDestinationCluster(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "cluster name is required"})
		return
	}

	log.G().WithFields(log.Fields{
		"name": name,
	}).Debug("getting destination cluster")

	// Get ArgoCD client
	argocdClient := argocd.GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	// Get cluster from ArgoCD
	cluster, err := clusterClient.Get(context.Background(), &clientCluster.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to get cluster %s: %v", name, err)
		c.JSON(404, gin.H{"error": fmt.Sprintf("Cluster %s not found", name)})
		return
	}

	// Get kubernetes client for the cluster
	destK8sClient, err := GetDestKubernetesClient(cluster)
	if err != nil {
		log.G().Errorf("Failed to get Kubernetes client for cluster %s: %v", name, err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to connect to cluster: %v", err)})
		return
	}

	// Get cluster version
	version, err := destK8sClient.Discovery().ServerVersion()
	if err != nil {
		log.G().Errorf("Failed to get server version: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get cluster version: %v", err)})
		return
	}

	total, readyNodes := countReadyNodes(destK8sClient)

	// Build response
	response := ClusterResponse{
		Name:        cluster.Name,
		Environment: cluster.Labels["environment"],
		Status:      getClusterStatus(destK8sClient),
		Provider:    cluster.Labels["vendor"],
		Version: VersionInfo{
			Kubernetes: version.GitVersion,
			Platform:   getPlatformVersion(version, cluster.Labels["vendor"]),
		},
		NodeCount:     total,
		Region:        cluster.Labels["region"],
		ResourceQuota: getResourceQuota(*cluster),
		Health: HealthStatus{
			Status:  cluster.Info.ConnectionState.Status,
			Message: cluster.Info.ConnectionState.Message,
		},
		Nodes: NodeStatus{
			Ready: readyNodes,
			Total: total,
		},
		NetworkPolicy:     true,
		IngressController: getIngressController(cluster.Labels["vendor"]),
		LastUpdated:       cluster.Info.ConnectionState.ModifiedAt.String(),
		ConsoleUrl:        getConsoleURL(*cluster),
		Monitoring:        getMonitoringInfo(*cluster),
		Builtin:           cluster.Labels["builtin"] == "true",
		Labels:            cluster.Labels,
	}

	c.JSON(200, response)
}
