package handler

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/store"
)

var (
	DefaultApplicationSetGeneratorInterval int64 = 20

	//go:embed assets/cluster_res_readme.md
	clusterResReadmeTpl []byte
)

type (
	ProjectCreateOptions struct {
		CloneOpts       *git.CloneOptions
		ProjectName     string
		DestKubeServer  string
		DestKubeContext string
		DryRun          bool
		AddCmd          argocd.AddClusterCmd
		Labels          map[string]string
		Annotations     map[string]string
	}

	ProjectDeleteOptions struct {
		CloneOpts   *git.CloneOptions
		ProjectName string
	}

	ProjectListOptions struct {
		CloneOpts *git.CloneOptions
		Out       io.Writer
	}

	GenerateProjectOptions struct {
		Name               string
		Namespace          string
		DefaultDestServer  string
		DefaultDestContext string
		RepoURL            string
		Revision           string
		InstallationPath   string
		Labels             map[string]string
		Annotations        map[string]string
	}
)

func generateProjectManifests(o *GenerateProjectOptions) (projectYAML, appSetYAML, clusterResReadme, clusterResConfig []byte, err error) {
	project := &argocdv1alpha1.AppProject{
		TypeMeta: metav1.TypeMeta{
			Kind:       argocdv1alpha1.AppProjectSchemaGroupVersionKind.Kind,
			APIVersion: argocdv1alpha1.AppProjectSchemaGroupVersionKind.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name,
			Namespace: o.Namespace,
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave":     "-2",
				"argocd.argoproj.io/sync-options":  "PruneLast=true",
				store.Default.DestServerAnnotation: o.DefaultDestServer,
			},
		},
		Spec: argocdv1alpha1.AppProjectSpec{
			SourceRepos: []string{"*"},
			Destinations: []argocdv1alpha1.ApplicationDestination{
				{
					Server:    "*",
					Namespace: "*",
				},
			},
			Description: fmt.Sprintf("%s project", o.Name),
			ClusterResourceWhitelist: []metav1.GroupKind{
				{
					Group: "*",
					Kind:  "*",
				},
			},
			NamespaceResourceWhitelist: []metav1.GroupKind{
				{
					Group: "*",
					Kind:  "*",
				},
			},
		},
	}
	if projectYAML, err = yaml.Marshal(project); err != nil {
		err = fmt.Errorf("failed to marshal AppProject: %w", err)
		return
	}

	appSetYAML, err = createAppSet(&createAppSetOptions{
		name:                        o.Name,
		namespace:                   o.Namespace,
		appName:                     fmt.Sprintf("%s-{{ userGivenName }}", o.Name),
		appNamespace:                o.Namespace,
		appProject:                  o.Name,
		repoURL:                     "{{ srcRepoURL }}",
		srcPath:                     "{{ srcPath }}",
		revision:                    "{{ srcTargetRevision }}",
		destServer:                  "{{ destServer }}",
		destNamespace:               "{{ destNamespace }}",
		prune:                       true,
		preserveResourcesOnDeletion: false,
		appLabels:                   getDefaultAppLabels(o.Labels),
		appAnnotations:              o.Annotations,
		generators: []argocdv1alpha1.ApplicationSetGenerator{
			{
				Git: &argocdv1alpha1.GitGenerator{
					RepoURL:  o.RepoURL,
					Revision: o.Revision,
					Files: []argocdv1alpha1.GitFileGeneratorItem{
						{
							Path: path.Join(o.InstallationPath, store.Default.AppsDir, "**", o.Name, "config.json"),
						},
					},
					RequeueAfterSeconds: &DefaultApplicationSetGeneratorInterval,
				},
			},
			{
				Git: &argocdv1alpha1.GitGenerator{
					RepoURL:  o.RepoURL,
					Revision: o.Revision,
					Files: []argocdv1alpha1.GitFileGeneratorItem{
						{
							Path: path.Join(o.InstallationPath, store.Default.AppsDir, "**", o.Name, "config_dir.json"),
						},
					},
					RequeueAfterSeconds: &DefaultApplicationSetGeneratorInterval,
					Template: argocdv1alpha1.ApplicationSetTemplate{
						Spec: argocdv1alpha1.ApplicationSpec{
							Source: &argocdv1alpha1.ApplicationSource{
								Directory: &argocdv1alpha1.ApplicationSourceDirectory{
									Recurse: true,
									Exclude: "{{ exclude }}",
									Include: "{{ include }}",
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to marshal ApplicationSet: %w", err)
		return
	}

	clusterResReadme = []byte(strings.ReplaceAll(string(clusterResReadmeTpl), "{CLUSTER}", o.DefaultDestServer))

	clusterResConfig, err = json.Marshal(&application.ClusterResConfig{Name: o.DefaultDestContext, Server: o.DefaultDestServer})
	if err != nil {
		err = fmt.Errorf("failed to create cluster resources config: %w", err)
		return
	}

	return
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

func getDefaultAppLabels(labels map[string]string) map[string]string {
	res := map[string]string{
		store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
		store.Default.LabelKeyAppName:      "{{ appName }}",
	}
	for k, v := range labels {
		res[k] = v
	}

	return res
}
