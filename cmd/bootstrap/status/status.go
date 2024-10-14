package status

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show bootstrap status",
		Run: func(cmd *cobra.Command, args []string) {
			kubeconfig := viper.GetString("kubernetes.kubeconfig.path")
			if kubeconfig == "" {
				fmt.Println("Error: kubernetes.kubeconfig.path is not set in the config file")
				return
			}

			config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				fmt.Printf("Error building kubeconfig: %v\n", err)
				return
			}

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				fmt.Printf("Error creating Kubernetes client: %v\n", err)
				return
			}

			k8sHealth := checkKubernetesHealth(clientset)

			argocdHealth := checkArgoCDHealth(clientset)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "COMPONENT\tSTATUS\tDETAILS")
			_, _ = fmt.Fprintf(w, "Kubernetes\t%s\t%s\n", k8sHealth.status, k8sHealth.details)
			_, _ = fmt.Fprintf(w, "ArgoCD\t%s\t%s\n", argocdHealth.status, argocdHealth.details)
			_ = w.Flush()
		},
	}
}

type healthStatus struct {
	status  string
	details string
}

func checkKubernetesHealth(clientset *kubernetes.Clientset) healthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return healthStatus{"Unhealthy", fmt.Sprintf("Error: %v", err)}
	}

	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return healthStatus{"Healthy", "Unable to fetch version"}
	}

	return healthStatus{"Healthy", fmt.Sprintf("Version: %s", version.GitVersion)}
}

func checkArgoCDHealth(clientset *kubernetes.Clientset) healthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deployment, err := clientset.AppsV1().Deployments("argocd").Get(ctx, "argocd-server", metav1.GetOptions{})
	if err != nil {
		return healthStatus{"Unhealthy", fmt.Sprintf("Error: %v", err)}
	}

	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		return healthStatus{"Healthy", fmt.Sprintf("Ready replicas: %d/%d", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)}
	}

	return healthStatus{"Unhealthy", fmt.Sprintf("Ready replicas: %d/%d", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)}
}
