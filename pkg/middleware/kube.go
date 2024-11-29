package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

// KubeFactoryMiddleware injects a kube factory into the context
func KubeFactoryMiddleware() gin.HandlerFunc {
	// Create factory and clients once when middleware is initialized
	factory := kube.NewFactory()
	restConfig, err := factory.ToRESTConfig()
	if err != nil {
		log.G().Fatalf("Failed to get REST config: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		log.G().Fatalf("Failed to create dynamic client: %v", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		log.G().Fatalf("Failed to create discovery client: %v", err)
	}

	return func(c *gin.Context) {
		c.Set("kubeFactory", factory)
		c.Set("dynamicClient", dynamicClient)
		c.Set("discoveryClient", discoveryClient)
		c.Next()
	}
}
