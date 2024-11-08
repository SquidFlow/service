package kube

import (
	"fmt"
	"os"
	"path/filepath"

	argocclient "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// defaultKubeconfig returns the default kubeconfig path
func defaultKubeconfig() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// isRunningInCluster checks if code is running in a kubernetes cluster
func isRunningInCluster() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}

// NewClient creates a new kubernetes clientset
func NewClient() (*kubernetes.Clientset, error) {
	if isRunningInCluster() {
		return createInClusterClient()
	}
	return createOutOfClusterClient(defaultKubeconfig())
}

func createInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster client: %w", err)
	}
	return clientset, nil
}

func createOutOfClusterClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("kubeconfig path is empty and not running in cluster")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create out-of-cluster client: %w", err)
	}

	return clientset, nil
}

// NewArgoCdClient creates a new ArgoCD client
func NewArgoCdClient() (*argocclient.ArgoprojV1alpha1Client, error) {
	if isRunningInCluster() {
		return createClusterInArgoCdClient()
	}
	return createOutOfClusterArgoCDClient(defaultKubeconfig())
}

func createClusterInArgoCdClient() (*argocclient.ArgoprojV1alpha1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientSet, err := argocclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster ArgoCD client: %w", err)
	}
	return clientSet, nil
}

func createOutOfClusterArgoCDClient(kubeconfigPath string) (*argocclient.ArgoprojV1alpha1Client, error) {
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("kubeconfig path is empty and not running in cluster")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	clientSet, err := argocclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create out-of-cluster ArgoCD client: %w", err)
	}

	return clientSet, nil
}
