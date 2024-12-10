package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"

	clusterpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/gin-gonic/gin"

	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/types"
)

// ClusterRegister creates a new destination cluster
// Note: for self-signed CA certificate, this API just forward the request to argocd-server
// you should confirm the CA has been mount to argocd-server pod
// otherwise, the API will reject your request with CA related error
func ClusterRegister(c *gin.Context) {
	var req types.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	log.G().WithFields(log.Fields{
		"name": req.Name,
		"env":  req.Env,
	}).Debug("user input create destination cluster")

	// Note: this api will create the cluster in argo-cd db without using argocd api so that
	// the argocd-server cache will not be tracked
	cls, err := argocd.RegisterCluster2ArgoCd(req.Name, req.Env, req.KubeConfig, req.Labels)
	if err != nil {
		if strings.Contains(err.Error(), "existing cluster") {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Cluster %s already exists", req.Name)})
			return
		}
		log.G().Errorf("Failed to create cluster: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create cluster: %v", err)})
		return
	}

	// parse the kubeConfig
	c.JSON(201, gin.H{
		"message": fmt.Sprintf("destination cluster: %s created successfully", req.Name),
		"cluster": cls,
	})
}

// ClusterDeregister deletes a destination cluster by name
func ClusterDeregister(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "cluster name is required"})
		return
	}

	log.G().WithFields(log.Fields{
		"name": name,
	}).Debug("deleting destination cluster")

	err := argocd.DeregisterCluster2ArgoCd(name)
	if err != nil {
		log.G().Errorf("Failed to delete cluster %s: %v", name, err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete cluster: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Destination cluster %s deleted successfully", name),
	})
}

// ClusterGet handles the GET request for a single cluster
func ClusterGet(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "cluster name is required"})
		return
	}

	log.G().WithFields(log.Fields{
		"name": name,
	}).Debug("getting destination cluster")

	cluster, err := argocd.GetCluster(name)
	if err != nil {
		log.G().Errorf("failed to get cluster %s: %v", name, err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get cluster: %v", err)})
		return
	}

	// Get kubernetes client for the cluster
	destK8sClient, err := GetDestKubernetesClient(cluster)
	if err != nil {
		log.G().Errorf("failed to get Kubernetes client for cluster %s: %v", name, err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to connect to cluster: %v", err)})
		return
	}

	// Get cluster version
	version, err := destK8sClient.Discovery().ServerVersion()
	if err != nil {
		log.G().Errorf("failed to get server version: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get cluster version: %v", err)})
		return
	}

	total, readyNodes := countReadyNodes(destK8sClient)

	// Build response
	response := &types.ClusterResponse{
		Name:        cluster.Name,
		Environment: cluster.Labels["environment"],
		Status:      getClusterStatus(destK8sClient),
		Provider:    cluster.Labels["vendor"],
		Version: types.VersionInfo{
			Kubernetes: version.GitVersion,
			Platform:   getPlatformVersion(version, cluster.Labels["vendor"]),
		},
		NodeCount:     total,
		Region:        cluster.Labels["region"],
		ResourceQuota: getResourceQuota(*cluster),
		Health: types.HealthStatus{
			Status:  cluster.Info.ConnectionState.Status,
			Message: cluster.Info.ConnectionState.Message,
		},
		Nodes: types.NodeStatus{
			Ready: readyNodes,
			Total: total,
		},
		NetworkPolicy:     true,
		IngressController: getIngressController(cluster.Labels["vendor"]),
		LastUpdated:       cluster.Info.ConnectionState.ModifiedAt.String(),
		ConsoleUrl:        getConsoleURL(*cluster),
		Monitoring:        getMonitoringInfo(*cluster),
		Builtin:           cluster.Labels["builtin"] == "true",
		Labels:            cluster.Labels,
	}

	c.JSON(200, response)
}

// ClusterList handles the GET request for listing clusters
func ClusterList(c *gin.Context) {
	clusterList, err := argocd.ListClusters()
	if err != nil {
		log.G().Errorf("failed to list clusters: %v", err)
		c.JSON(500, gin.H{"error": "failed to list clusters"})
		return
	}

	response := &types.ClusterListResponse{
		Success: true,
		Total:   len(clusterList.Items),
		Message: "success",
		Items:   []types.ClusterResponse{},
		Error:   "",
	}

	for _, cluster := range clusterList.Items {
		destK8sClient, err := GetDestKubernetesClient(&cluster)
		if err != nil {
			log.G().Warnf("Failed to get Kubernetes client with TLS for cluster %s: %v", cluster.Name, err)
			continue
		}

		version, err := destK8sClient.Discovery().ServerVersion()
		if err != nil {
			log.G().Errorf("Failed to get server version: %v", err)
			continue
		}

		total, readyNodes := countReadyNodes(destK8sClient)
		clusterInfo := types.ClusterResponse{
			Name:        cluster.Name,
			Environment: cluster.Annotations["squidflow.github.io/cluster-env"],
			Status:      getClusterStatus(destK8sClient),
			Provider:    cluster.Annotations["squidflow.github.io/cluster-vendor"],
			Version: types.VersionInfo{
				Kubernetes: version.GitVersion,
				Platform:   getPlatformVersion(version, cluster.Annotations["squidflow.github.io/cluster-vendor"]),
			},
			NodeCount:     total,
			Region:        "hk",
			ResourceQuota: getResourceQuota(cluster),
			Health: types.HealthStatus{
				Status:  cluster.Info.ConnectionState.Status,
				Message: cluster.Info.ConnectionState.Message,
			},
			Nodes: types.NodeStatus{
				Ready: readyNodes,
				Total: total,
			},
			NetworkPolicy:     true, // This should be determined based on cluster configuration
			IngressController: getIngressController(cluster.Labels["vendor"]),
			LastUpdated:       time.Now().String(),
			ConsoleUrl:        getConsoleURL(cluster),
			Monitoring:        getMonitoringInfo(cluster),
			Builtin:           cluster.Labels["builtin"] == "true",
			Labels:            cluster.Labels,
		}
		response.Items = append(response.Items, clusterInfo)
	}

	c.JSON(200, response)
}

func ClusterUpdate(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "cluster name is required"})
		return
	}

	var req types.UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	log.G().WithFields(log.Fields{
		"name": name,
		"env":  req.Env,
	}).Debug("updating destination cluster")

	// Get ArgoCD client
	argocdClient := argocd.GetArgoServerClient()
	closer, clusterClient := argocdClient.NewClusterClientOrDie()
	defer closer.Close()

	// First get existing cluster
	existingCluster, err := clusterClient.Get(context.Background(), &clusterpkg.ClusterQuery{
		Name: name,
	})
	if err != nil {
		log.G().Errorf("Failed to get cluster %s: %v", name, err)
		c.JSON(404, gin.H{"error": fmt.Sprintf("Cluster %s not found", name)})
		return
	}

	var restConfig *rest.Config
	if req.KubeConfig != "" {
		// Only process kubeconfig if it's provided
		kubeconfigBytes, err := base64.StdEncoding.DecodeString(req.KubeConfig)
		if err != nil {
			log.G().Errorf("Failed to decode kubeConfig: %v", err)
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to decode kubeconfig: %v", err)})
			return
		}

		// Parse kubeconfig
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
		if err != nil {
			log.G().Errorf("Failed to parse kubeConfig: %v", err)
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to parse kubeconfig: %v", err)})
			return
		}
	}

	// Merge default labels with custom labels
	annotations := map[string]string{
		"squidflow.github.io/cluster-env":    req.Env,
		"squidflow.github.io/cluster-vendor": "aliyun",
		"squidflow.github.io/cluster-name":   name,
	}

	// Add custom labels
	for k, v := range req.Labels {
		annotations[k] = v
	}

	updatedCluster := existingCluster
	updatedCluster.Annotations = annotations

	// Only update server config if kubeconfig was provided
	if restConfig != nil {
		updatedCluster.Server = restConfig.Host
		updatedCluster.Config.TLSClientConfig = argoappv1.TLSClientConfig{
			Insecure:   restConfig.TLSClientConfig.Insecure,
			ServerName: restConfig.TLSClientConfig.ServerName,
			CAData:     restConfig.TLSClientConfig.CAData,
			CertData:   restConfig.TLSClientConfig.CertData,
			KeyData:    restConfig.TLSClientConfig.KeyData,
		}
	}

	updatedCluster.Config.TLSClientConfig.Insecure = true
	updatedCluster.Config.TLSClientConfig.CAData = nil

	result, err := clusterClient.Update(context.Background(), &clusterpkg.ClusterUpdateRequest{
		Cluster: updatedCluster,
	})
	if err != nil {
		log.G().Errorf("Failed to update cluster %s: %v", name, err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update cluster: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Destination cluster %s updated successfully", name),
		"cluster": result,
	})
}

// GetDestKubernetesClient returns a Kubernetes clientset with TLS configuration
// improve: URIToSecretName
func GetDestKubernetesClient(argocdCluster *argoappv1.Cluster) (kubernetes.Interface, error) {
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
			clusterSecret = secret.Data["config"]
			break
		}
	}

	if clusterSecret == nil {
		return nil, fmt.Errorf("no matching secret found for cluster %s", argocdCluster.Name)
	}

	// Parse the TLS config
	var tlsConfig types.TLSClientConfig
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

func countReadyNodes(destCluster kubernetes.Interface) (total, ready int) {
	nodes, err := destCluster.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.G().Errorf("Failed to list nodes: %v", err)
		return 0, 0
	}
	readyNodes := 0
	for _, node := range nodes.Items {
		if node.Status.Conditions[0].Status == corev1.ConditionTrue {
			readyNodes++
		}
	}
	return len(nodes.Items), readyNodes
}

// TODO: need to implement
func getIngressController(vendor string) string {
	log.G().WithFields(log.Fields{
		"vendor": vendor,
	}).Debugf("Getting ingress controller for vendor")

	return "nginx"
}

// TODO: need implement
func getClusterStatus(destCluster kubernetes.Interface) []types.ComponentStatus {
	cs, err := destCluster.CoreV1().ComponentStatuses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []types.ComponentStatus{{
			Name:    "cluster",
			Status:  "degraded",
			Message: "Failed to get component status",
			Error:   err.Error(),
		}}
	}

	var components []types.ComponentStatus
	for _, component := range cs.Items {
		status := types.ComponentStatus{
			Name: component.Name,
		}

		status.Status = "Healthy"
		status.Message = "ok"

		for _, condition := range component.Conditions {
			if condition.Status != "True" {
				status.Status = "Unhealthy"
				status.Message = condition.Message
				if condition.Error != "" {
					status.Error = condition.Error
				}
				break
			}
		}

		components = append(components, status)
	}

	return components
}

func getPlatformVersion(version *version.Info, vendor string) string {
	log.G().Infof("Version: %v", version)
	if vendor == "aws" {
		return version.Platform
	}
	return version.GitVersion
}

func getResourceQuota(cluster argoappv1.Cluster) types.ResourceQuota {
	log.G().WithFields(log.Fields{
		"cluster name": cluster.Name,
	}).Debugf("Getting resource quota for cluster")

	return types.ResourceQuota{
		CPU:       "64 cores",
		Memory:    "256Gi",
		Storage:   "5000Gi",
		PVCs:      "50",
		NodePorts: "20",
	}
}

func getConsoleURL(cluster argoappv1.Cluster) string {
	log.G().WithFields(log.Fields{
		"cluster name": cluster.Name,
	}).Debugf("Getting console URL for cluster")

	return "https://console.aws.amazon.com/eks/home?region=us-west-2#/clusters/" + cluster.Name
}

func getMonitoringInfo(cluster argoappv1.Cluster) types.MonitoringInfo {
	log.G().WithFields(log.Fields{
		"cluster":        cluster.Name,
		"cluster labels": cluster.Labels,
	}).Debugf("Getting monitoring info for cluster")

	return types.MonitoringInfo{
		Prometheus:   true,
		Grafana:      true,
		AlertManager: true,
		URLs: &types.MonitoringURLs{
			Prometheus:   "http://prometheus.argo-cd.svc.cluster.local",
			Grafana:      "http://grafana.argo-cd.svc.cluster.local",
			AlertManager: "http://alertmanager.argo-cd.svc.cluster.local",
		},
	}
}
