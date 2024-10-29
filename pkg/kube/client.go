package kube

import (
	"fmt"

	argocclient "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned/typed/application/v1alpha1"
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

func NewArgoCdClient() (*argocclient.ArgoprojV1alpha1Client, error) {
	mode := viper.GetString("kubernetes.mode")
	switch mode {
	case "incluster":
		return createClusterInArgoCdClient()
	case "kubeconfig":
		return createClusterArgoCDClient(viper.GetString("kubernetes.kubeconfig.path"))
	default:
		return nil, fmt.Errorf("invalid config")
	}
}

func createClusterInArgoCdClient() (*argocclient.ArgoprojV1alpha1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientSet, err := argocclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientSet, nil
}

func createClusterArgoCDClient(filePath string) (*argocclient.ArgoprojV1alpha1Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filePath)
	if err != nil {
		panic(err)
	}

	clientSet, err := argocclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientSet, nil
}
