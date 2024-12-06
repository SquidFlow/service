package argocd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/squidflow/service/pkg/log"
)

func RegisterCluster2ArgoCd(name, env, kubeconfig string, labels map[string]string) (*argoappv1.Cluster, error) {
	// parse the kubeConfig
	kubconfigWithoutBase64, err := base64.StdEncoding.DecodeString(kubeconfig)
	if err != nil {
		log.G().Errorf("Failed to decode kubeConfig: %v", err)
		return nil, err
	}

	// 1. get rest config from kubeconfig
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubconfigWithoutBase64)
	if err != nil {
		log.G().Errorf("failed to parse kubeConfig: %v", err)
		return nil, err
	}

	annotations := map[string]string{
		"squidflow.github.io/cluster-env":    env,
		"squidflow.github.io/cluster-vendor": "aliyun",
		"squidflow.github.io/cluster-name":   name,
	}

	// detect the vendor
	createClusterReq := NewArgoCdClusterCreateReq(
		name,
		[]string{},
		true,
		restConfig,
		"",
		nil,
		nil,
		labels,
		annotations,
	)
	log.G().WithFields(log.Fields{
		"cluster":           createClusterReq.Name,
		"cluster_labels":    createClusterReq.Labels,
		"cluster_annotaion": createClusterReq.Annotations,
	}).Debug("argocd request for create destination cluster")

	// Get ArgoCD client
	argocdClient := GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	cls, err := clusterClient.Create(context.Background(), &clusterpkg.ClusterCreateRequest{
		Cluster: createClusterReq,
	})
	if err != nil {
		log.G().Errorf("failed to create cluster in argo-cd db: %v", err)
		return nil, err
	}

	return cls, nil
}

func DeregisterCluster2ArgoCd(name string) error {
	// Get ArgoCD client
	argocdClient := GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	// First check if cluster exists
	cluster, err := clusterClient.Get(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("failed to get cluster %s: %v", name, err)
		return fmt.Errorf("cluster %s not found", name)
	}

	log.G().WithFields(log.Fields{
		"name":   name,
		"server": cluster.Server,
	}).Debug("found cluster, proceeding with deletion")

	// Delete the cluster
	_, err = clusterClient.Delete(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to delete cluster %s: %v", name, err)
		return fmt.Errorf("failed to delete cluster: %v", err)
	}

	return nil
}

func GetCluster(name string) (*argoappv1.Cluster, error) {
	// Get ArgoCD client
	argocdClient := GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			log.G().Errorf(err.Error())
		}
	}(closer)

	// Get cluster from ArgoCD
	cluster, err := clusterClient.Get(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to get cluster %s: %v", name, err)
		return nil, fmt.Errorf("cluster %s not found", name)
	}

	return cluster, nil
}

func ListClusters() (*argoappv1.ClusterList, error) {
	argocdClient := GetArgoServerClient()
	closer, clsClient := argocdClient.NewClusterClientOrDie()
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			log.G().Errorf(err.Error())
		}
	}(closer)

	clusterList, err := clsClient.List(context.Background(), &clusterpkg.ClusterQuery{})
	if err != nil {
		log.G().Errorf("failed to list clusters: %v", err)
		return nil, fmt.Errorf("failed to list clusters: %v", err)
	}

	log.G().WithFields(log.Fields{
		"cluster count": len(clusterList.Items),
	}).Debug("list destination cluster found clusters count")

	return clusterList, nil
}

func NewArgoCdClusterCreateReq(name string,
	namespaces []string,
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

	// if insecure mode is set, ensure CA data is cleared
	if tlsClientConfig.Insecure {
		tlsClientConfig.CAData = nil
		log.G().Warn("using insecure TLS configuration for cluster connection")
	}

	cls := argoappv1.Cluster{
		Server:           conf.Host,
		Name:             name,
		Namespaces:       namespaces,
		ClusterResources: clusterResources,
		Config: argoappv1.ClusterConfig{
			TLSClientConfig: tlsClientConfig,
			// AWSAuthConfig:      awsAuthConf,
			// ExecProviderConfig: execProviderConf,
		},
		Labels:      labels,
		Annotations: annotations,
	}

	// Bearer token will preferentially be used for auth if present,
	// Even in presence of key/cert credentials
	// So set bearer token only if the key/cert data is absent
	if len(tlsClientConfig.CertData) == 0 || len(tlsClientConfig.KeyData) == 0 {
		cls.Config.BearerToken = managerBearerToken
	}

	return &cls
}
