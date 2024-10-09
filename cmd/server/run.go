package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/h4-poc/service/pkg/autopilot"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the server",
	Long:  `Run the Application API server`,
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Panic: %v", r)
		}
	}()

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	_, err = getKubernetesClientset(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	_, err = getSCMClient(config)
	if err != nil {
		log.Fatalf("Failed to create SCM client: %v", err)
	}

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/applications", createApplication)
		v1.GET("/applications", listApplications)
		v1.GET("/applications/:appName", getApplication)
		v1.PUT("/applications/:appName", updateApplication)
		v1.DELETE("/applications/:appName", deleteApplication)
	}

	serverAddr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Starting server on %s", serverAddr)
	r.Run(serverAddr)
}

func loadConfig() (*Config, error) {
	return nil, nil
}

func getKubernetesClientset(config *Config) (interface{}, error) {
	return nil, nil
}

func getSCMClient(config *Config) (interface{}, error) {
	return nil, nil
}

func createApplication(c *gin.Context) {
	var newApp autopilot.Application
	if err := c.BindJSON(&newApp); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 验证必要的字段
	if newApp.AppName == "" || newApp.Repo == "" {
		c.JSON(400, gin.H{"error": "app-name and repo are required"})
		return
	}

	// 将应用程序数据写入本地文件
	if err := autopilot.WriteApplicationToFile(newApp); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save application data"})
		return
	}

	c.JSON(201, newApp)
}

// Application 结构体定义
type Application struct {
	AppName     string `json:"app-name"`
	Repo        string `json:"repo"`
	WaitTimeout string `json:"wait-timeout"`
}

func listApplications(c *gin.Context) {
	// 实现列出应用程序的逻辑
}

func getApplication(c *gin.Context) {
	// 实现获取单个应用程序的逻辑
}

func updateApplication(c *gin.Context) {
	// 实现更新应用程序的逻辑
}

func deleteApplication(c *gin.Context) {
	// 实现删除应用程序的逻辑
}

// Config 结构体定义
type Config struct {
	Server struct {
		Port int
	}
	// 其他配置字段...
}
