package handler

import (
	"context"
	"encoding/base64"
	"fmt"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/log"
)

// UpdateClusterRequest represents the request body for cluster update
type UpdateClusterRequest struct {
	Env        string            `json:"env" binding:"required,oneof=DEV SIT UAT PRD"`
	KubeConfig string            `json:"kubeconfig,omitempty"` // base64 encoded, optional
	Labels     map[string]string `json:"labels,omitempty"`     // custom labels
}

func UpdateDestinationCluster(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "cluster name is required"})
		return
	}

	var req UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	log.G().WithFields(log.Fields{
		"name": name,
		"env":  req.Env,
	}).Debug("updating destination cluster")

	// Get ArgoCD client
	argocdClient := argocd.GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	// First get existing cluster
	existingCluster, err := clusterClient.Get(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to get cluster %s: %v", name, err)
		c.JSON(404, gin.H{"error": fmt.Sprintf("Cluster %s not found", name)})
		return
	}

	var restConfig *rest.Config
	if req.KubeConfig != "" {
		// Only process kubeconfig if it's provided
		kubeconfigBytes, err := base64.StdEncoding.DecodeString(req.KubeConfig)
		if err != nil {
			log.G().Errorf("Failed to decode kubeConfig: %v", err)
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to decode kubeconfig: %v", err)})
			return
		}

		// Parse kubeconfig
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
		if err != nil {
			log.G().Errorf("Failed to parse kubeConfig: %v", err)
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to parse kubeconfig: %v", err)})
			return
		}
	}

	// Merge default labels with custom labels
	annotations := map[string]string{
		"squidflow.github.io/cluster-env":    req.Env,
		"squidflow.github.io/cluster-vendor": "aliyun",
		"squidflow.github.io/cluster-name":   name,
	}

	// Add custom labels
	for k, v := range req.Labels {
		annotations[k] = v
	}

	updatedCluster := existingCluster
	updatedCluster.Annotations = annotations

	// Only update server config if kubeconfig was provided
	if restConfig != nil {
		updatedCluster.Server = restConfig.Host
		updatedCluster.Config.TLSClientConfig = argoappv1.TLSClientConfig{
			Insecure:   restConfig.TLSClientConfig.Insecure,
			ServerName: restConfig.TLSClientConfig.ServerName,
			CAData:     restConfig.TLSClientConfig.CAData,
			CertData:   restConfig.TLSClientConfig.CertData,
			KeyData:    restConfig.TLSClientConfig.KeyData,
		}
	}

	result, err := clusterClient.Update(context.Background(), &clusterpkg.ClusterUpdateRequest{
		Cluster: updatedCluster,
	})
	if err != nil {
		log.G().Errorf("Failed to update cluster %s: %v", name, err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update cluster: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Destination cluster %s updated successfully", name),
		"cluster": result,
	})
}
