package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// validateClusterRequest validates the cluster creation request
func validateClusterRequest(req *CreateClusterRequest) error {
	// 验证环境名称
	validEnvs := map[string]bool{
		"SIT": true,
		"UAT": true,
		"PRD": true,
	}
	if !validEnvs[req.Environment] {
		return fmt.Errorf("invalid environment: %s", req.Environment)
	}

	// 验证供应商
	validProviders := map[string]bool{
		"GKE": true,
		"OCP": true,
		"AKS": true,
		"EKS": true,
	}
	if !validProviders[req.Provider] {
		return fmt.Errorf("invalid provider: %s", req.Provider)
	}

	// 验证资源配额
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

	// 验证请求
	if err := validateClusterRequest(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Validation failed: %v", err)})
		return
	}

	// 获取 Kubernetes 客户端
	clientset, err := getKubernetesClient()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get kubernetes client: %v", err)})
		return
	}

	// 获取集群版本信息
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get cluster version: %v", err)})
		return
	}

	// 获取节点信息
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
		NodeCount:        len(nodes.Items),
		Region:          req.Region,
		ResourceQuota:   req.ResourceQuota,
		Health:          getClusterHealth(clientset),
		Nodes: NodeStatus{
			Ready: readyNodes,
			Total: len(nodes.Items),
		},
		NetworkPolicy:    req.NetworkPolicy,
		IngressController: req.IngressController,
		LastUpdated:      time.Now().UTC().Format(time.RFC3339),
		ConsoleURL:       req.ConsoleURL,
		Monitoring:       req.Monitoring,
	}

	// TODO: 将集群信息保存到持久化存储中
	// 这里应该实现将集群信息保存到数据库或其他存储中的逻辑

	c.JSON(201, cluster)
}
