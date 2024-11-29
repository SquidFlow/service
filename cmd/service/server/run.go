package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	clusterclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/discovery"

	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/config"
	"github.com/squidflow/service/pkg/handler"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/middleware"
	"github.com/squidflow/service/pkg/store"
)

func NewRunCommand() *cobra.Command {
	var (
		configFile     string
		kubeconfigPath string
	)

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the server",
		Long:  `Run the Application API server`,
		Run: func(cmd *cobra.Command, args []string) {
			runServer(cmd, args)
		},
	}

	runCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to config file")
	runCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "Path to kubeconfig file (default is $HOME/.kube/config)")

	err := runCmd.MarkFlagRequired("config")
	if err != nil {
		panic(err)
	}

	return runCmd
}

func runServer(cmd *cobra.Command, args []string) {
	defer func() {
		if r := recover(); r != nil {
			log.G().Fatalf("Panic recovered: %v", r)
		}
	}()

	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		log.G().Fatalf("Failed to get config file: %v", err)
	}

	_, err = config.ParseConfig(configFile)
	if err != nil {
		log.G().Fatalf("Failed to load config: %v", err)
	}

	// Set gin mode based on environment
	if viper.GetString("env") == "production" || os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
		logger := log.G()
		if err := logger.Configure(); err != nil {
			log.G().Fatalf("Failed to configure logger: %v", err)
		}
		log.G().Info("Running in production mode")
	} else {
		gin.SetMode(gin.DebugMode)
		logger := log.G()
		if err := logger.Configure(); err != nil {
			log.G().Fatalf("Failed to configure logger: %v", err)
		}
		log.G().Info("Running in development mode")
	}

	// Display version information
	log.G().WithFields(log.Fields{
		"Version":    store.Get().Version.Version,
		"GitCommit":  store.Get().Version.GitCommit,
		"BuildTime":  store.Get().Version.BuildDate,
		"GoCompiler": store.Get().Version.GoCompiler,
		"GoVersion":  store.Get().Version.GoVersion,
		"Platform":   store.Get().Version.Platform,
	}).Info("Starting service")

	// Create kubernetes clients
	factory := kube.NewFactory()
	restConfig, err := factory.ToRESTConfig()
	if err != nil {
		log.G().Fatalf("Failed to get REST config: %v", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		log.G().Fatalf("Failed to create discovery client: %v", err)
	}

	// Check required CRDs
	log.G().Info("Checking required CRDs...")
	if err := checkRequiredCRDs(discoveryClient); err != nil {
		log.G().Fatalf("CRD check failed: %v", err)
	}

	// connect to ArgoCD API server
	argocdClient := argocd.GetArgoServerClient()
	if argocdClient == nil {
		log.G().Fatalf("Failed to create argocd client")
	}

	closer, clsClient, err := argocdClient.NewClusterClient()
	if err != nil {
		log.G().Fatalf("Failed to create cluster client: %v", err)
	}
	defer closer.Close()

	err = listDestinationCluster(context.Background(), clsClient)
	if err != nil {
		log.G().Fatalf("Failed to list destination clusters: %v", err)
	}

	//TODO: check gitOps repo
	r := setupRouter()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("server.port")),
		Handler: r,
	}

	go func() {
		log.G().Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.G().Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.G().Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.G().Errorf("Server forced to shutdown: %v", err)
	}

	log.G().Info("Server exiting")
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.KubeFactoryMiddleware())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/healthz", handler.Healthz)
	}

	// app code
	{
		v1.GET("/appcode", handler.ListAppCode)
	}

	// the target cluster of argo application
	// cluster name is required, immutable, unique
	clusters := v1.Group("/destinationCluster")
	{
		clusters.POST("", handler.CreateDestinationCluster)
		clusters.GET("", handler.ListDestinationCluster)
		clusters.GET("/:name", handler.GetDestinationCluster)
		clusters.DELETE("/:name", handler.DeleteDestinationCluster)
		clusters.PATCH("/:name", handler.UpdateDestinationCluster)
	}

	// real api, to manage the lifecycle of ArgoApplication
	applications := v1.Group("/deploy/argocdapplications")
	{
		applications.POST("", handler.CreateArgoApplication)
		applications.GET("", handler.ListArgoApplications)
		applications.POST("/sync", handler.SyncArgoApplication)
		applications.POST("/dryrun", handler.DryRunArgoApplications)
		applications.POST("/dryruntemplate", handler.ApplicationTemplateDryRun)
		applications.POST("/validate", handler.ValidateTemplate)

		app := applications.Group("/:name")
		{
			app.GET("", handler.DescribeArgoApplication)
			app.PATCH("", handler.UpdateArgoApplication)
			app.DELETE("", handler.DeleteArgoApplication)
		}
	}

	// one tenant : one ArgoCD Project
	tenants := v1.Group("/tenants")
	{
		tenants.POST("", handler.CreateTenant)
		tenants.GET("", handler.ListTenants)
		tenantsOne := tenants.Group("/:name")
		{
			tenantsOne.DELETE("", handler.DeleteTenant)
			tenantsOne.GET("", handler.DescribeTenant)
		}
	}

	// integrated with ExternalSecrets
	security := v1.Group("/security")
	{
		secretStore := security.Group("/externalsecrets/secretstore")
		{
			secretStore.POST("", handler.CreateSecretStore)
			secretStore.GET("", handler.ListSecretStore)
			secretStoreOne := secretStore.Group("/:id")
			{
				secretStoreOne.GET("", handler.DescribeSecretStore)
				secretStoreOne.PATCH("", handler.UpdateSecretStore)
				secretStoreOne.DELETE("", handler.DeleteSecretStore)
			}
		}
	}

	return r
}

// listDestinationCluster
func listDestinationCluster(ctx context.Context, clsClient clusterpkg.ClusterServiceClient) error {
	// list cluster
	clusterList, err := clsClient.List(context.Background(), &clusterclient.ClusterQuery{})
	if err != nil {
		log.G().Error("Failed to list clusters: %v", err)
		return err
	}

	log.G().Info("Available clusters:")
	log.G().Info(strings.Repeat("-", 80))
	log.G().Info(fmt.Sprintf("%-60s\t%-30s\t%-10s", "Name", "Server", "Status"))

	for _, cls := range clusterList.Items {
		status := cls.Info.ConnectionState.Status
		log.G().Info(fmt.Sprintf("%-60s\t%-30s\t%-10s",
			cls.Name,
			cls.Server,
			status))
	}
	log.G().Info(strings.Repeat("-", 80))

	return nil
}

func checkRequiredCRDs(discoveryClient *discovery.DiscoveryClient) error {
	requiredCRDs := []struct {
		group    string
		resource string
	}{
		{"argoproj.io", "applications"},
		{"argoproj.io", "applicationsets"},
		{"argocd-addon.github.com", "applicationtemplates"},
		{"argoproj.io", "appprojects"},
		{"external-secrets.io", "clusterexternalsecrets"},
		{"external-secrets.io", "clustersecretstores"},
		{"external-secrets.io", "externalsecrets"},
		{"external-secrets.io", "secretstores"},
	}

	resources, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return fmt.Errorf("failed to get server resources: %w", err)
	}

	missingCRDs := []string{}

	for _, crd := range requiredCRDs {
		found := false
		for _, list := range resources {
			if !strings.Contains(list.GroupVersion, crd.group) {
				continue
			}
			for _, r := range list.APIResources {
				if r.Name == crd.resource {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			missingCRDs = append(missingCRDs, fmt.Sprintf("%s.%s", crd.resource, crd.group))
		}
	}

	if len(missingCRDs) > 0 {
		return fmt.Errorf("required CRDs are not installed: %s", strings.Join(missingCRDs, ", "))
	}

	log.G().Info("All required CRDs are installed")
	return nil
}
