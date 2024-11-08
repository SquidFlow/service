package handler

import (
	"fmt"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/log"
)

// CreateClusterRequest represents the request body for cluster creation
type CreateClusterRequest struct {
	KubeConfig string `json:"kubeconfig" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Env        string `json:"env" binding:"required,oneof=SIT UAT PRD"`
	Vendor     string `json:"vendor" binding:"required,oneof=GKE OCP AKS EKS"`
}

// CreateDestinationCluster creates a new destination cluster
func CreateDestinationCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// parse the kubeConfig
	rawConfig, err := clientcmd.Load([]byte(req.KubeConfig))
	if err != nil {
		log.G().Errorf("Failed to parse kubeconfig: %v", err)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to parse kubeconfig: %v", err)})
		return
	}

	// Validate kubeconfig: should contain only one context
	if len(rawConfig.Contexts) != 1 {
		log.G().Errorf("Kubeconfig should contain exactly one context, found %d", len(rawConfig.Contexts))
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("Kubeconfig should contain exactly one context, found %d", len(rawConfig.Contexts)),
		})
		return
	}

	// Get the only context
	var contextName string
	var context *api.Context
	for name, ctx := range rawConfig.Contexts {
		contextName = name
		context = ctx
		break
	}

	// Validate cluster exists in clusters section
	cluster, exists := rawConfig.Clusters[context.Cluster]
	if !exists {
		log.G().Errorf("Cluster %s not found in kubeconfig", context.Cluster)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Cluster %s not found in kubeconfig", context.Cluster)})
		return
	}

	// Create rest config from the context
	restConfig, err := clientcmd.NewNonInteractiveClientConfig(
		*rawConfig,
		contextName,
		&clientcmd.ConfigOverrides{},
		nil,
	).ClientConfig()
	if err != nil {
		log.G().Errorf("Failed to create rest config: %v", err)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to create rest config: %v", err)})
		return
	}

	// Create cluster object
	clst := clusterpkg.ClusterCreateRequest{
		Cluster: &v1alpha1.Cluster{
			Server: cluster.Server, // Use server from cluster config
			Name:   req.Name,
			Config: v1alpha1.ClusterConfig{
				TLSClientConfig: v1alpha1.TLSClientConfig{
					Insecure:   cluster.InsecureSkipTLSVerify,
					ServerName: cluster.TLSServerName,
					CAData:     cluster.CertificateAuthorityData,
					CertData:   restConfig.TLSClientConfig.CertData,
					KeyData:    restConfig.TLSClientConfig.KeyData,
				},
				BearerToken: restConfig.BearerToken,
			},
			Labels: map[string]string{
				"environment": req.Env,
				"vendor":      req.Vendor,
				"context":     contextName,
			},
		},
	}

	// Get ArgoCD client
	argocdClient := argocd.GetArgoServerClient()
	if argocdClient == nil {
		log.G().Error("Failed to get argocd client")
		c.JSON(500, gin.H{"error": "Failed to get argocd client"})
		return
	}

	// Create cluster in ArgoCD
	closer, clusterClient, err := argocdClient.NewClusterClient()
	if err != nil {
		log.G().Errorf("Failed to create cluster client: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create cluster client: %v", err)})
		return
	}
	defer io.Close(closer)

	_, err = clusterClient.Create(c.Request.Context(), &clst)
	if err != nil {
		log.G().Errorf("Failed to create cluster: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create cluster: %v", err)})
		return
	}

	c.JSON(201, gin.H{
		"message": fmt.Sprintf("Cluster %s created successfully", req.Name),
		"cluster": clst,
	})
}
