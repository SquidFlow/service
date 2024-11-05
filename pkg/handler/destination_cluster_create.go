package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	validEnvs = map[string]bool{
		"SIT": true,
		"UAT": true,
		"PRD": true,
	}
	validProviders = map[string]bool{
		"GKE": true,
		"OCP": true,
		"AKS": true,
		"EKS": true,
	}
)

// validateClusterRequest validates the cluster creation request
func validateClusterRequest(req *CreateClusterRequest) error {
	if !validEnvs[req.Environment] {
		return fmt.Errorf("invalid environment: %s", req.Environment)
	}

	if !validProviders[req.Provider] {
		return fmt.Errorf("invalid provider: %s", req.Provider)
	}

	if req.ResourceQuota.CPU == "" || req.ResourceQuota.Memory == "" || req.ResourceQuota.Storage == "" {
		return fmt.Errorf("resource quota must be specified")
	}

	return nil
}

// CreateDestinationCluster creates a new destination cluster
func CreateDestinationCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	if err := validateClusterRequest(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Validation failed: %v", err)})
		return
	}

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
		Name:        req.Name,
		Environment: req.Environment,
		Status:      "active",
		Provider:    req.Provider,
		Version: ClusterVersion{
			Kubernetes: version.GitVersion,
			Platform:   fmt.Sprintf("%s %s", req.Provider, version.GitVersion),
		},
		NodeCount:     len(nodes.Items),
		Region:        req.Region,
		ResourceQuota: req.ResourceQuota,
		Health:        getClusterHealth(clientset),
		Nodes: NodeStatus{
			Ready: readyNodes,
			Total: len(nodes.Items),
		},
		NetworkPolicy:     req.NetworkPolicy,
		IngressController: req.IngressController,
		LastUpdated:       time.Now().UTC().Format(time.RFC3339),
		ConsoleURL:        req.ConsoleURL,
		Monitoring:        req.Monitoring,
	}

	c.JSON(201, cluster)
}
