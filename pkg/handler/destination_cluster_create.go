package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/log"
)

// CreateClusterRequest represents the request body for cluster creation
type CreateClusterRequest struct {
	Name       string            `json:"name" binding:"required"`
	Env        string            `json:"env" binding:"required,oneof=DEV SIT UAT PRD"`
	KubeConfig string            `json:"kubeconfig" binding:"required"` // with base64 encoding
	Labels     map[string]string `json:"labels,omitempty"`              // custom labels
}

// CreateDestinationCluster creates a new destination cluster
func CreateDestinationCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	log.G().WithFields(log.Fields{
		"name": req.Name,
		"env":  req.Env,
	}).Debug("user input create destination cluster")

	// parse the kubeConfig
	kubconfigWithoutBase64, err := base64.StdEncoding.DecodeString(req.KubeConfig)
	if err != nil {
		log.G().Errorf("Failed to decode kubeConfig: %v", err)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to decode kubeconfig: %v", err)})
		return
	}

	// Parse kubeconfig into config object
	config, err := clientcmd.Load(kubconfigWithoutBase64)
	if err != nil {
		log.G().Errorf("Failed to load kubeconfig: %v", err)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to load kubeconfig: %v", err)})
		return
	}

	// Log important kubeconfig information
	log.G().WithFields(log.Fields{
		"current-context": config.CurrentContext,
		"contexts":        len(config.Contexts),
		"clusters":        len(config.Clusters),
		"users":           len(config.AuthInfos),
	}).Debug("Parsed kubeconfig details")

	// If you want to log specific cluster info
	if cluster, exists := config.Clusters[config.CurrentContext]; exists {
		log.G().WithFields(log.Fields{
			"server":                   cluster.Server,
			"insecure-skip-tls-verify": cluster.InsecureSkipTLSVerify,
			"certificate-authority":    len(cluster.CertificateAuthority) > 0,
		}).Debug("Current cluster details")
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubconfigWithoutBase64)
	if err != nil {
		log.G().Errorf("Failed to parse kubeConfig: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to parse kubeconfig: %v", err)})
		return
	}

	// Merge default labels with custom labels
	annotations := map[string]string{
		"h4-poc.github.io/cluster-env":    req.Env,
		"h4-poc.github.io/cluster-vendor": "aliyun",
		"h4-poc.github.io/cluster-name":   req.Name,
	}

	// Add custom labels
	labels := make(map[string]string)
	for k, v := range req.Labels {
		labels[k] = v
	}

	// detect the vendor
	createClusterReq := newArgoCdClusterCreateReq(
		req.Name,
		[]string{},
		true,
		restConfig,
		"",
		nil,
		nil,
		labels,
		annotations, // Use merged annotations
	)
	log.G().WithFields(log.Fields{
		"cluster":           createClusterReq.Name,
		"cluster_labels":    createClusterReq.Labels,
		"cluster_annotaion": createClusterReq.Annotations,
	}).Debug("argocd request for create destination cluster")

	// 2. argoCd cluster client
	argocdClient := argocd.GetArgoServerClient()
	closer, clsClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	cls, err := clsClient.Create(context.Background(), &clusterpkg.ClusterCreateRequest{
		Cluster: createClusterReq,
	})
	if err != nil {
		if strings.Contains(err.Error(), "existing cluster") {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Cluster %s already exists", req.Name)})
			return
		}
		log.G().Errorf("Failed to create cluster: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create cluster: %v", err)})
		return
	}

	// parse the kubeConfig
	c.JSON(201, gin.H{
		"message": fmt.Sprintf("destination cluster: %s created successfully", req.Name),
		"cluster": cls,
	})
}
