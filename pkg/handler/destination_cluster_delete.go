package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/log"
)

// DeleteDestinationCluster deletes a destination cluster by name
func DeleteDestinationCluster(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "cluster name is required"})
		return
	}

	log.G().WithFields(log.Fields{
		"name": name,
	}).Debug("deleting destination cluster")

	// Get ArgoCD client
	argocdClient := argocd.GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	// First check if cluster exists
	cluster, err := clusterClient.Get(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to get cluster %s: %v", name, err)
		c.JSON(404, gin.H{"error": fmt.Sprintf("Cluster %s not found", name)})
		return
	}

	log.G().WithFields(log.Fields{
		"name":   name,
		"server": cluster.Server,
	}).Debug("found cluster, proceeding with deletion")

	// Delete the cluster
	_, err = clusterClient.Delete(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to delete cluster %s: %v", name, err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete cluster: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Destination cluster %s deleted successfully", name),
	})
}
