package commands

import (
	"context"
	_ "embed"
	"fmt"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/util"
)

var (
	die = util.Die

	prepareRepo = func(ctx context.Context, cloneOpts *git.CloneOptions, projectName string) (git.Repository, fs.FS, error) {
		log.G().WithFields(log.Fields{
			"repoURL":  cloneOpts.URL(),
			"revision": cloneOpts.Revision(),
			"forWrite": cloneOpts.CloneForWrite,
		}).Debug("starting with options: ")

		// clone repo
		log.G().Infof("cloning git repository: %s", cloneOpts.URL())
		r, repofs, err := getRepo(ctx, cloneOpts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed cloning the repository: %w", err)
		}

		root := repofs.Root()
		log.G().Infof("using revision: \"%s\", installation path: \"%s\"", cloneOpts.Revision(), root)
		if !repofs.ExistsOrDie(store.Default.BootsrtrapDir) {
			return nil, nil, fmt.Errorf("bootstrap directory not found, please execute `repo bootstrap` command")
		}

		if projectName != "" {
			projExists := repofs.ExistsOrDie(repofs.Join(store.Default.ProjectsDir, projectName+".yaml"))
			if !projExists {
				return nil, nil, fmt.Errorf(util.Doc(fmt.Sprintf("project '%[1]s' not found, please execute `<BIN> project create %[1]s`", projectName)))
			}
		}

		log.G().Debug("repository is ok")

		return r, repofs, nil
	}
)

type createAppOptions struct {
	name        string
	namespace   string
	repoURL     string
	revision    string
	srcPath     string
	destServer  string
	noFinalizer bool
	labels      map[string]string
}

func createApp(opts *createAppOptions) ([]byte, error) {
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

type createAppSetOptions struct {
	name                        string
	namespace                   string
	appName                     string
	appNamespace                string
	appProject                  string
	repoURL                     string
	revision                    string
	srcPath                     string
	destServer                  string
	destNamespace               string
	prune                       bool
	preserveResourcesOnDeletion bool
	appLabels                   map[string]string
	appAnnotations              map[string]string
	generators                  []argocdv1alpha1.ApplicationSetGenerator
}

func createAppSet(o *createAppSetOptions) ([]byte, error) {
	if o.destServer == "" {
		o.destServer = store.Default.DestServer
	}

	if o.appProject == "" {
		o.appProject = "default"
	}

	if o.appLabels == nil {
		// default labels
		o.appLabels = map[string]string{
			store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
			"app.kubernetes.io/name":           o.appName,
		}
	}

	appSet := &argocdv1alpha1.ApplicationSet{
		TypeMeta: metav1.TypeMeta{
			// do not use argocdv1alpha1.ApplicationSetSchemaGroupVersionKind.Kind because it is "Applicationset" - noticed the lowercase "s"
			Kind:       "ApplicationSet",
			APIVersion: argocdv1alpha1.ApplicationSetSchemaGroupVersionKind.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.name,
			Namespace: o.namespace,
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": "0",
			},
		},
		Spec: argocdv1alpha1.ApplicationSetSpec{
			Generators: o.generators,
			Template: argocdv1alpha1.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: argocdv1alpha1.ApplicationSetTemplateMeta{
					Namespace:   o.appNamespace,
					Name:        o.appName,
					Labels:      o.appLabels,
					Annotations: o.appAnnotations,
				},
				Spec: argocdv1alpha1.ApplicationSpec{
					Project: o.appProject,
					Source: &argocdv1alpha1.ApplicationSource{
						RepoURL:        o.repoURL,
						Path:           o.srcPath,
						TargetRevision: o.revision,
					},
					Destination: argocdv1alpha1.ApplicationDestination{
						Server:    o.destServer,
						Namespace: o.destNamespace,
					},
					SyncPolicy: &argocdv1alpha1.SyncPolicy{
						Automated: &argocdv1alpha1.SyncPolicyAutomated{
							SelfHeal:   true,
							Prune:      o.prune,
							AllowEmpty: true,
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
			},
			SyncPolicy: &argocdv1alpha1.ApplicationSetSyncPolicy{
				PreserveResourcesOnDeletion: o.preserveResourcesOnDeletion,
			},
		},
	}

	return yaml.Marshal(appSet)
}

var getInstallationNamespace = func(repofs fs.FS) (string, error) {
	path := repofs.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName+".yaml")
	a := &argocdv1alpha1.Application{}
	if err := repofs.ReadYamls(path, a); err != nil {
		return "", fmt.Errorf("failed to unmarshal namespace: %w", err)
	}

	return a.Spec.Destination.Namespace, nil
}
