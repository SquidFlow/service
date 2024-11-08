package handler

import (
	"context"
	"encoding/base64"
	"fmt"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/log"
)

// CreateClusterRequest represents the request body for cluster creation
type CreateClusterRequest struct {
	Env        string `json:"env" binding:"required,oneof=SIT UAT PRD"`
	KubeConfig string `json:"kubeconfig" binding:"required"` // with base64 encoding
}

// CreateDestinationCluster creates a new destination cluster
func CreateDestinationCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	log.G().WithFields(log.Fields{
		"env":        req.Env,
		"kubeconfig": req.KubeConfig,
	}).Debug("user input create destination cluster")

	// parse the kubeConfig
	kubconfigWithoutBase64, err := base64.StdEncoding.DecodeString(req.KubeConfig)
	if err != nil {
		log.G().Errorf("Failed to decode kubeConfig: %v", err)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to decode kubeconfig: %v", err)})
		return
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubconfigWithoutBase64)
	if err != nil {
		log.G().Errorf("Failed to parse kubeConfig: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to parse kubeconfig: %v", err)})
		return
	}

	// detect the vendor
	createClusterReq := NewCluster(
		restConfig.Host, []string{},
		true,
		restConfig,
		"",
		nil,
		nil,
		map[string]string{"env": req.Env, "vendor": "aliyun"},
		nil)
	log.G().WithFields(log.Fields{
		"cluster":           createClusterReq.Name,
		"cluster label":     createClusterReq.Labels,
		"cluster annotaion": createClusterReq.Annotations,
	}).Debug("argocd request for create destination cluster")

	// 2. argoCd cluster client
	argocdClient := argocd.GetArgoServerClient()
	closer, clsClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	cls, err := clsClient.Create(context.Background(), &clusterpkg.ClusterCreateRequest{
		Cluster: createClusterReq,
	})
	if err != nil {
		log.G().Errorf("Failed to create cluster: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create cluster: %v", err)})
		return
	}

	// parse the kubeConfig
	c.JSON(201, gin.H{
		"message": fmt.Sprintf("destination cluster: %s created successfully", cls.Name),
		"cluster": cls,
	})
}
