package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	clientCluster "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"

	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/log"
)

// ListDestinationCluster handles the GET request for listing clusters
func ListDestinationCluster(c *gin.Context) {
	argocdClient := argocd.GetArgoServerClient()
	closer1, clsClient := argocdClient.NewClusterClientOrDie()
	defer closer1.Close()

	clusterList, err := clsClient.List(context.Background(), &clientCluster.ClusterQuery{})
	if err != nil {
		log.G().Fatalf("Failed to list clusters: %v", err)
	}

	c.JSON(200, clusterList)
}
