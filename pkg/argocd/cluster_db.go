package argocd

import (
	"context"
	"fmt"
	"os"

	"github.com/argoproj/argo-cd/v2/util/db"
	"github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
)

func NewArgoCDDB(ctx context.Context) (db.ArgoDB, error) {
	log.G().Infof("new argo-cd db: namespace")

	namespace := os.Getenv("ARGOCD_NAMESPACE")
	if namespace == "" {
		namespace = "argocd"
	}

	// Create kubernetes client to get secrets
	factory := kube.NewFactory()
	k8sClient, err := factory.KubernetesClientSet()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	settingsMgr := settings.NewSettingsManager(ctx, k8sClient, namespace)

	return db.NewDB(namespace, settingsMgr, k8sClient), nil
}
