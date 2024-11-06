package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/config"
	"github.com/h4-poc/service/pkg/handler"
	"github.com/h4-poc/service/pkg/log"
)

func NewRunCommand() *cobra.Command {
	var configFile string
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the server",
		Long:  `Run the Application API server`,
		Run:   runServer,
	}
	runCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to config file")
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
	r.Use(corsMiddleware())
	r.Use(requestIDMiddleware())

	v1 := r.Group("/api/v1")

	{
		r.GET("/healthz", handler.Healthz)
	}

	// the target cluster of argo application
	clusters := v1.Group("/destinationCluster")
	{
		clusters.GET("", handler.ListDestinationCluster)
		clusters.POST("", handler.CreateDestinationCluster)
		clusters.PATCH("/:name", handler.UpdateDestinationCluster)
	}

	// save the template resource to improve the user experience
	// actually. to want handle the manifest repo
	templates := v1.Group("/applications/templates")
	{
		templates.POST("", handler.CreateApplicationTemplate)
		templates.GET("", handler.ListApplicationTemplate)
		templates.POST("/validate", handler.ValidateApplicationTemplate)
		templateInstance := templates.Group("/:template_id")
		{
			templateInstance.PATCH("", handler.UpdateApplicationTemplate)
			templateInstance.DELETE("", handler.UpdateApplicationTemplate)
		}
	}

	// real api, to manage the lifecycle of ArgoApplication
	applications := v1.Group("/deploy/argocdapplications")
	{
		applications.POST("", handler.CreateArgoApplication)
		applications.GET("", handler.ListArgoApplications)
		applications.POST("/sync", handler.SyncArgoApplication)
		applications.POST("/dryrun", handler.DryRunArgoApplications)

		app := applications.Group("/:appName")
		{
			app.GET("", handler.DescribeArgoApplications)
			app.PATCH("", handler.UpdateArgoApplication)
			app.DELETE("", handler.DeleteArgoApplication)
		}
	}

	// one tenant : one ArgoCD Project
	tenants := v1.Group("/tenants")
	{
		tenants.POST("", handler.CreateProject)
		tenants.GET("", handler.ListProjects)
		tenantsOne := tenants.Group("/:tenantName")
		{
			tenantsOne.GET("", handler.ListProjects)
			tenantsOne.DELETE("", handler.DeleteProject)
		}
	}

	// integrated with ExternalSecrets
	security := v1.Group("/security")
	{
		secretStore := security.Group("/externalsecrets/secretstore")
		{
			secretStore.POST("", handler.CreateSecretStore)
			secretStore.GET("", handler.ListSecretStore)
			secretStoreOne := secretStore.Group("/:name")
			{
				secretStoreOne.PATCH("", handler.UpdateSecretStore)
				secretStoreOne.DELETE("", handler.DeleteSecretStore)
			}
		}
	}

	return r
}

// corsMiddleware to handle CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// requestIDMiddleware injects a request ID into the context
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID to generate request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}
