package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDestinationCluster handles the GET request for listing clusters
func ListDestinationCluster(c *gin.Context) {
	clientset, err := getKubernetesClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get kubernetes client: %v", err)})
		return
	}

	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get cluster version: %v", err)})
		return
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get nodes: %v", err)})
		return
	}

	readyNodes := 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				readyNodes++
				break
			}
		}
	}

	cluster := ClusterInfo{
		Name:        "current-cluster", // 这应该从配置或存储中获取
		Environment: "default",         // 这应该从配置或存储中获取
		Status:      "active",
		Provider:    "unknown", // 这应该从配置或存储中获取
		Version: ClusterVersion{
			Kubernetes: version.GitVersion,
			Platform:   fmt.Sprintf("Unknown %s", version.GitVersion),
		},
		NodeCount: len(nodes.Items),
		Region:    "unknown", // 这应该从配置或存储中获取
		ResourceQuota: ResourceQuota{
			CPU:       "unlimited",
			Memory:    "unlimited",
			Storage:   "unlimited",
			PVCs:      "unlimited",
			NodePorts: "unlimited",
		},
		Health: getClusterHealth(clientset),
		Nodes: NodeStatus{
			Ready: readyNodes,
			Total: len(nodes.Items),
		},
		NetworkPolicy:     false,
		IngressController: "unknown",
		LastUpdated:       metav1.Now().Format(time.RFC3339),
		Monitoring: Monitoring{
			Prometheus:   false,
			Grafana:      false,
			Alertmanager: false,
		},
	}

	c.JSON(200, []ClusterInfo{cluster})
}
