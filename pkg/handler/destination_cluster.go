package handler

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

// CreateClusterRequest represents the request body for creating a new cluster
type CreateClusterRequest struct {
	Name              string        `json:"name" binding:"required"`
	Environment       string        `json:"environment" binding:"required"`
	Provider          string        `json:"provider" binding:"required"`
	Region            string        `json:"region" binding:"required"`
	ConsoleURL        string        `json:"consoleUrl,omitempty"`
	ResourceQuota     ResourceQuota `json:"resourceQuota"`
	NetworkPolicy     bool          `json:"networkPolicy"`
	IngressController string        `json:"ingressController"`
	Monitoring        Monitoring    `json:"monitoring"`
}

// getKubernetesClient returns a Kubernetes clientset
func getKubernetesClient() (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

// getClusterHealth checks the health status of a Kubernetes cluster
func getClusterHealth(clientset kubernetes.Interface) HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return HealthStatus{
			Status:  "Degraded",
			Message: fmt.Sprintf("API Server health check failed: %v", err),
		}
	}

	components, err := clientset.CoreV1().ComponentStatuses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return HealthStatus{
			Status:  "Warning",
			Message: fmt.Sprintf("Failed to check core components: %v", err),
		}
	}

	for _, component := range components.Items {
		for _, condition := range component.Conditions {
			if condition.Status != "True" {
				return HealthStatus{
					Status:  "Warning",
					Message: fmt.Sprintf("Component %s is unhealthy: %s", component.Name, condition.Message),
				}
			}
		}
	}

	return HealthStatus{
		Status:  "Healthy",
		Message: "All core components are healthy",
	}
}
