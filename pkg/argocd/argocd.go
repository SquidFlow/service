package argocd

import (
	"context"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argocdcs "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/log"
)

type (
	AddClusterCmd interface {
		Execute(ctx context.Context, clusterName string) error
	}

	addClusterImpl struct {
		cmd  *cobra.Command
		args []string
	}

	LoginOptions struct {
		Namespace   string
		Username    string
		Password    string
		KubeConfig  string
		KubeContext string
		Insecure    bool
	}
)

func (a *addClusterImpl) Execute(ctx context.Context, clusterName string) error {
	a.cmd.SetArgs(append(a.args, clusterName))
	return a.cmd.ExecuteContext(ctx)
}

// GetAppSyncWaitFunc returns a WaitFunc that will return true when the Application
// is in Sync + Healthy state, and at the specific revision (if supplied. If revision is "", no revision check is made)
func GetAppSyncWaitFunc(revision string, waitForCreation bool) kube.WaitFunc {
	return func(ctx context.Context, f kube.Factory, ns, name string) (bool, error) {
		rc, err := f.ToRESTConfig()
		if err != nil {
			return false, err
		}

		c, err := argocdcs.NewForConfig(rc)
		if err != nil {
			return false, err
		}

		app, err := c.ArgoprojV1alpha1().Applications(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			se, ok := err.(*kerrors.StatusError)
			if !waitForCreation || !ok || se.ErrStatus.Reason != metav1.StatusReasonNotFound {
				return false, err
			}

			return false, nil
		}

		synced := string(app.Status.Sync.Status) == string(v1alpha1.SyncStatusCodeSynced)
		healthy := app.Status.Health.Status == "Healthy"
		onRevision := true
		if revision != "" {
			onRevision = revision == app.Status.Sync.Revision
		}

		log.G().Debugf("Application found, Sync Status: %s, Health Status: %s, Revision: %s", app.Status.Sync.Status, app.Status.Health.Status, app.Status.Sync.Revision)
		return synced && healthy && onRevision, nil
	}
}
