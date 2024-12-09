package argocd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/squidflow/service/pkg/log"
)

const (
	AnnotationKeyName           = "squidflow.github.io/cluster-name"
	AnnotationKeyEnvironment    = "squidflow.github.io/cluster-env"
	AnnotationKeyTenant         = "squidflow.github.io/cluster-tenant"
	AnnotationKeyVendor         = "squidflow.github.io/cluster-vendor"
	AnnotationKeyDescription    = "squidflow.github.io/cluster-description"
	AnnotationKeyAppCode        = "squidflow.github.io/cluster-appcode"
	AnnotationKeyManaged        = "squidflow.github.io/managed"
	AnnotationKeyRegisterBy     = "squidflow.github.io/register-by"
	AnnotationKeyRegisterAt     = "squidflow.github.io/register-at"
	AnnotationKeyLastModifiedBy = "squidflow.github.io/last-modified-by"
	AnnotationKeyLastModifiedAt = "squidflow.github.io/last-modified-at"
)

func RegisterCluster2ArgoCd(name, env, kubeconfig string, ann map[string]string) (*argoappv1.Cluster, error) {
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

	if ann == nil {
		ann = map[string]string{}
	} else {
		ann[AnnotationKeyEnvironment] = env
		ann[AnnotationKeyVendor] = "aliyun"
		ann[AnnotationKeyName] = name
		ann[AnnotationKeyManaged] = "true"
		ann[AnnotationKeyRegisterBy] = "squidflow"
		ann[AnnotationKeyRegisterAt] = time.Now().Format(time.RFC3339)
	}

	labels := map[string]string{}
	labels[AnnotationKeyManaged] = "true"

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
		ann,
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
		if strings.Contains(err.Error(), "while trying to verify candidate authority certificate") {
			log.G().Warn("will force insert cluster to argocd-server cache, please confirm the CA has been mounted to argocd-server pod")
			log.G().Warn("the argocd-server cache will not be set, if you care about the cache, please update it again")
			argoDB, err := NewArgoCDDB(context.Background())
			if err != nil {
				log.G().Errorf("failed to create argo-cd db: %v", err)
				return nil, err
			}
			cls, err = argoDB.CreateCluster(context.Background(), createClusterReq)
			if err != nil {
				log.G().Errorf("failed to create cluster in argo-cd db: %v", err)
				return nil, err
			}
			log.G().WithFields(log.Fields{
				"cluster": cls.Name,
			}).Debug("cluster created in argo-cd db")
			return cls, nil
		}
		return nil, err
	}

	return cls, nil
}

func DeregisterCluster2ArgoCd(name string) error {
	argocdClient := GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

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

	_, err = clusterClient.Delete(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to delete cluster %s: %v", name, err)
		return fmt.Errorf("failed to delete cluster: %v", err)
	}

	return nil
}

func getClusterInfoFromAnnotations(annotations map[string]string) (environment, vendor string) {
	if annotations == nil {
		return "", ""
	}
	environment = annotations[AnnotationKeyEnvironment]
	vendor = annotations[AnnotationKeyVendor]
	return
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

	for i := range clusterList.Items {
		cluster := &clusterList.Items[i]
		env, vendor := getClusterInfoFromAnnotations(cluster.Annotations)

		if env != "" {
			if cluster.Annotations == nil {
				cluster.Annotations = make(map[string]string)
			}
			cluster.Annotations[AnnotationKeyEnvironment] = env
		}
		if vendor != "" {
			if cluster.Annotations == nil {
				cluster.Annotations = make(map[string]string)
			}
			cluster.Annotations[AnnotationKeyVendor] = vendor
		}
	}

	log.G().WithFields(log.Fields{
		"cluster count": len(clusterList.Items),
	}).Debug("list destination cluster found clusters count")

	return clusterList, nil
}

func GetCluster(name string) (*argoappv1.Cluster, error) {
	argocdClient := GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			log.G().Errorf(err.Error())
		}
	}(closer)

	cluster, err := clusterClient.Get(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to get cluster %s: %v", name, err)
		return nil, fmt.Errorf("cluster %s not found", name)
	}

	env, vendor := getClusterInfoFromAnnotations(cluster.Annotations)

	if env != "" {
		if cluster.Annotations == nil {
			cluster.Annotations = make(map[string]string)
		}
		cluster.Annotations[AnnotationKeyEnvironment] = env
	}
	if vendor != "" {
		if cluster.Annotations == nil {
			cluster.Annotations = make(map[string]string)
		}
		cluster.Annotations[AnnotationKeyVendor] = vendor
	}

	return cluster, nil
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
