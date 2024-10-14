package kube

import (
	"fmt"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient() (*kubernetes.Clientset, error) {
	mode := viper.GetString("kubernetes.mode")
	switch mode {
	case "incluster":
		return createInClusterClient()
	case "kubeconfig":
		return createClusterClient(viper.GetString("kubernetes.kubeconfig.path"))
	default:
		return nil, fmt.Errorf("invalid config")
	}
}

func createInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset, nil
}

func createClusterClient(filePath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filePath)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset, nil
}
