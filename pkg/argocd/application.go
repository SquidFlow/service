package argocd

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/squidflow/service/pkg/store"
)

type CreateAppOptions struct {
	name        string
	namespace   string
	repoURL     string
	revision    string
	srcPath     string
	destServer  string
	noFinalizer bool
	labels      map[string]string
}

func CreateApp(opts *CreateAppOptions) ([]byte, error) {
	if opts.destServer == "" {
		opts.destServer = store.Default.DestServer
	}

	app := &argocdv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       argocdv1alpha1.ApplicationSchemaGroupVersionKind.Kind,
			APIVersion: argocdv1alpha1.ApplicationSchemaGroupVersionKind.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.namespace,
			Name:      opts.name,
			Labels: map[string]string{
				store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
				"app.kubernetes.io/name":           opts.name,
			},
			Finalizers: []string{
				"resources-finalizer.argocd.argoproj.io",
			},
		},
		Spec: argocdv1alpha1.ApplicationSpec{
			Project: "default",
			Source: &argocdv1alpha1.ApplicationSource{
				RepoURL:        opts.repoURL,
				Path:           opts.srcPath,
				TargetRevision: opts.revision,
			},
			Destination: argocdv1alpha1.ApplicationDestination{
				Server:    opts.destServer,
				Namespace: opts.namespace,
			},
			SyncPolicy: &argocdv1alpha1.SyncPolicy{
				Automated: &argocdv1alpha1.SyncPolicyAutomated{
					SelfHeal:   true,
					Prune:      true,
					AllowEmpty: true,
				},
				SyncOptions: []string{
					"allowEmpty=true",
				},
			},
			IgnoreDifferences: []argocdv1alpha1.ResourceIgnoreDifferences{
				{
					Group: "argoproj.io",
					Kind:  "Application",
					JSONPointers: []string{
						"/status",
					},
				},
			},
		},
	}
	if opts.noFinalizer {
		app.ObjectMeta.Finalizers = []string{}
	}
	if len(opts.labels) > 0 {
		for k, v := range opts.labels {
			app.ObjectMeta.Labels[k] = v
		}
	}

	return yaml.Marshal(app)
}
