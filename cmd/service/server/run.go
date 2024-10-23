package server

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/h4-poc/service/pkg/config"
	"github.com/h4-poc/service/pkg/handler"
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
			log.Fatalf("Panic: %v", r)
		}
	}()

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf("Failed to get config file: %v", err)
	}

	_, err = config.ParseConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/applications", handler.CreateApplication)
		v1.GET("/applications", handler.ListApplications)
		v1.GET("/applications/:appName", handler.ListApplications)
		v1.PUT("/applications/:appName", handler.UpdateApplication)
		v1.DELETE("/applications", handler.DeleteApplication)
	}
	r.GET("/healthz", handler.Healthz)

	serverPort := viper.GetInt("server.port")
	serverAddr := fmt.Sprintf(":%d", serverPort)
	log.Printf("Starting server on %s", serverAddr)
	err = r.Run(serverAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
