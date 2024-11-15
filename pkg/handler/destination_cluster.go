package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/log"
)

// ClusterInfo represents a Kubernetes cluster's information
type ClusterInfo struct {
	Name              string         `json:"name"`
	Environment       string         `json:"environment"`
	Status            string         `json:"status"`
	Provider          string         `json:"provider"`
	Version           ClusterVersion `json:"version"`
	NodeCount         int            `json:"nodeCount"`
	Region            string         `json:"region"`
	ResourceQuota     ResourceQuota  `json:"resourceQuota"`
	Health            HealthStatus   `json:"health"`
	Nodes             NodeStatus     `json:"nodes"`
	NetworkPolicy     bool           `json:"networkPolicy"`
	IngressController string         `json:"ingressController"`
	LastUpdated       string         `json:"lastUpdated"`
	ConsoleURL        string         `json:"consoleUrl,omitempty"`
	Monitoring        Monitoring     `json:"monitoring"`
	Builtin           bool           `json:"builtin,omitempty"`
}

type ClusterVersion struct {
	Kubernetes string `json:"kubernetes"`
	Platform   string `json:"platform"`
}

type Monitoring struct {
	Prometheus   bool        `json:"prometheus"`
	Grafana      bool        `json:"grafana"`
	Alertmanager bool        `json:"alertmanager"`
	URLs         MonitorURLs `json:"urls,omitempty"`
}

type MonitorURLs struct {
	Prometheus   string `json:"prometheus,omitempty"`
	Grafana      string `json:"grafana,omitempty"`
	Alertmanager string `json:"alertmanager,omitempty"`
}

// TLSClientConfig represents the structure of the config data in the secret
type TLSClientConfig struct {
	TLSClientConfig struct {
		Insecure bool   `json:"insecure"`
		CertData string `json:"certData"`
		KeyData  string `json:"keyData"`
		CAData   string `json:"caData"`
	} `json:"tlsClientConfig"`
}

// getKubernetesClient returns a Kubernetes clientset
func getKubernetesClient() (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

// getClusterHealth checks the health status of a Kubernetes cluster
func getClusterHealth(clientset kubernetes.Interface) HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return HealthStatus{
			Status:  "Degraded",
			Message: fmt.Sprintf("API Server health check failed: %v", err),
		}
	}

	components, err := clientset.CoreV1().ComponentStatuses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return HealthStatus{
			Status:  "Warning",
			Message: fmt.Sprintf("Failed to check core components: %v", err),
		}
	}

	for _, component := range components.Items {
		for _, condition := range component.Conditions {
			if condition.Status != "True" {
				return HealthStatus{
					Status:  "Warning",
					Message: fmt.Sprintf("Component %s is unhealthy: %s", component.Name, condition.Message),
				}
			}
		}
	}

	return HealthStatus{
		Status:  "Healthy",
		Message: "All core components are healthy",
	}
}

// GetDestKubernetesClient returns a Kubernetes clientset with TLS configuration
// improve: URIToSecretName
func GetDestKubernetesClient(argocdCluster *argoappv1.Cluster) (kubernetes.Interface, error) {
	log.G().Debugf("Getting kubernetes client for cluster: %s", argocdCluster.Name)

	// Create kubernetes client to get secrets
	factory := kube.NewFactory()
	k8sClient, err := factory.KubernetesClientSet()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	if argocdCluster.Name == "in-cluster" {
		return k8sClient, nil
	}

	// List secrets with the ArgoCD cluster label
	secrets, err := k8sClient.CoreV1().Secrets("argocd").List(context.Background(), metav1.ListOptions{
		LabelSelector: "argocd.argoproj.io/secret-type=cluster",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	// Find matching secret
	var clusterSecret []byte
	for _, secret := range secrets.Items {
		if string(secret.Data["name"]) == argocdCluster.Name {
			log.G().Debugf("found matching secret: %s", secret.Name)
			clusterSecret = secret.Data["config"]
			break
		}
	}

	if clusterSecret == nil {
		return nil, fmt.Errorf("no matching secret found for cluster %s", argocdCluster.Name)
	}

	// Parse the TLS config
	var tlsConfig TLSClientConfig
	if err := json.Unmarshal(clusterSecret, &tlsConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TLS config: %w", err)
	}

	// Create REST config
	restConfig := &rest.Config{
		Host: argocdCluster.Server,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: tlsConfig.TLSClientConfig.Insecure,
		},
	}

	// Decode and set certificate data
	if tlsConfig.TLSClientConfig.CAData != "" {
		caData, err := base64.StdEncoding.DecodeString(tlsConfig.TLSClientConfig.CAData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode CA data: %w", err)
		}
		restConfig.TLSClientConfig.CAData = caData
	}
	if tlsConfig.TLSClientConfig.CertData != "" {
		certData, err := base64.StdEncoding.DecodeString(tlsConfig.TLSClientConfig.CertData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode cert data: %w", err)
		}
		restConfig.TLSClientConfig.CertData = certData
	}
	if tlsConfig.TLSClientConfig.KeyData != "" {
		keyData, err := base64.StdEncoding.DecodeString(tlsConfig.TLSClientConfig.KeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode key data: %w", err)
		}
		restConfig.TLSClientConfig.KeyData = keyData
	}

	// Create and return kubernetes client
	return kubernetes.NewForConfig(restConfig)
}

func newArgoCdClusterCreateReq(name string, namespaces []string,
	clusterResources bool,
	conf *rest.Config,
	managerBearerToken string,
	awsAuthConf *argoappv1.AWSAuthConfig,
	execProviderConf *argoappv1.ExecProviderConfig,
	labels, annotations map[string]string,
) *argoappv1.Cluster {
	tlsClientConfig := argoappv1.TLSClientConfig{
		Insecure:   conf.TLSClientConfig.Insecure,
		ServerName: conf.TLSClientConfig.ServerName,
		CAData:     conf.TLSClientConfig.CAData,
		CertData:   conf.TLSClientConfig.CertData,
		KeyData:    conf.TLSClientConfig.KeyData,
	}

	clst := argoappv1.Cluster{
		Server:           conf.Host,
		Name:             name,
		Namespaces:       namespaces,
		ClusterResources: clusterResources,
		Config: argoappv1.ClusterConfig{
			TLSClientConfig:    tlsClientConfig,
			AWSAuthConfig:      awsAuthConf,
			ExecProviderConfig: execProviderConf,
		},
		Labels:      labels,
		Annotations: annotations,
	}

	// Bearer token will preferentially be used for auth if present,
	// Even in presence of key/cert credentials
	// So set bearer token only if the key/cert data is absent
	if len(tlsClientConfig.CertData) == 0 || len(tlsClientConfig.KeyData) == 0 {
		clst.Config.BearerToken = managerBearerToken
	}

	return &clst
}
