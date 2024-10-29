package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/kyaml/filesys"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argocdsettings "github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/ghodss/yaml"
	"github.com/go-git/go-billy/v5/memfs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kusttypes "sigs.k8s.io/kustomize/api/types"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
)

func NewBootstrapCmd() *cobra.Command {
	var (
		appSpecifier               = ""
		dryRun                     = false
		hidePassword               = false
		insecure                   = false
		recover                    = false
		installationMode           = installationModeNormal
		cloneOpts                  = &git.CloneOptions{}
		namespaceLabels            = map[string]string{}
		applicationRepoRemoteURL   = ""
		applicationRepoAccessToken = ""
		kubeConfigPath             = ""
	)

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "bootstrap the the platform",
		Example: util.Doc(`
# Install argo-cd on the current kubernetes context in the argocd namespace
# and persists the bootstrap manifests to the root of gitops repository

	supervisor <BIN> bootstrap
`),
		PreRun: func(cmd *cobra.Command, args []string) {
			// in recover mode we don't want to commit anything
			if recover {
				cloneOpts.CloneForWrite = false
				cloneOpts.CreateIfNotExist = false
			}

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cloneOpts = &git.CloneOptions{
				FS:               fs.Create(memfs.New()),
				CreateIfNotExist: true,
				CloneForWrite:    true,
				Repo:             applicationRepoRemoteURL,
				Provider:         "github",
				Auth: git.Auth{
					Password: applicationRepoAccessToken,
				},
			}
			cloneOpts.Parse()

			return RunRepoBootstrap(cmd.Context(), &RepoBootstrapOptions{
				AppSpecifier:     appSpecifier,
				InstallationMode: installationMode,
				Namespace:        "",
				KubeConfig:       kubeConfigPath,
				KubeContextName:  "",
				DryRun:           dryRun,
				HidePassword:     hidePassword,
				Insecure:         insecure,
				Recover:          recover,
				Timeout:          util.MustParseDuration("180s"),
				KubeFactory:      kube.NewFactory(),
				CloneOptions:     cloneOpts,
				NamespaceLabels:  namespaceLabels,
			})
		},
	}

	log.Debugf("start clone options %s", applicationRepoRemoteURL)

	cmd.Flags().StringVar(&appSpecifier, "app", "", "Application specifier")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run")
	cmd.Flags().BoolVar(&hidePassword, "hide-password", false, "Hide password")
	cmd.Flags().BoolVar(&insecure, "insecure", false, "Insecure")
	cmd.Flags().BoolVar(&recover, "recover", false, "Recover")
	cmd.Flags().StringVar(&installationMode, "installation-mode", installationModeNormal, "Installation mode")
	cmd.Flags().StringToStringVar(&namespaceLabels, "namespace-labels", nil, "Namespace labels")
	cmd.Flags().StringVar(&applicationRepoAccessToken, "git-token", "", "git token")
	cmd.Flags().StringVar(&applicationRepoRemoteURL, "repo", "", "application repo")
	cmd.Flags().StringVar(&kubeConfigPath, "kube-config", "", "kube config path")

	return cmd
}

const (
	installationModeFlat   = "flat"
	installationModeNormal = "normal"
)

// used for mocking
var (
	//go:embed assets/cluster_res_readme.md
	clusterResReadmeTpl []byte

	//go:embed assets/projects_readme.md
	projectReadme []byte

	//go:embed assets/apps_readme.md
	appsReadme []byte

	exit                                         = os.Exit
	currentKubeContext                           = kube.CurrentContext
	runKustomizeBuild                            = application.GenerateManifests
	DefaultApplicationSetGeneratorInterval int64 = 20

	getRepo = func(ctx context.Context, cloneOpts *git.CloneOptions) (git.Repository, fs.FS, error) {
		return cloneOpts.GetRepo(ctx)
	}
)

type (
	RepoBootstrapOptions struct {
		AppSpecifier        string
		InstallationMode    string
		Namespace           string
		KubeConfig          string
		KubeContextName     string
		DryRun              bool
		HidePassword        bool
		Insecure            bool
		Recover             bool
		Timeout             time.Duration
		KubeFactory         kube.Factory
		CloneOptions        *git.CloneOptions
		ArgoCDLabels        map[string]string
		BootstrapAppsLabels map[string]string
		NamespaceLabels     map[string]string
	}

	RepoUninstallOptions struct {
		Namespace       string
		KubeContextName string
		Timeout         time.Duration
		CloneOptions    *git.CloneOptions
		KubeFactory     kube.Factory
		Force           bool
		FastExit        bool
	}

	bootstrapManifests struct {
		bootstrapApp           []byte
		rootApp                []byte
		clusterResAppSet       []byte
		clusterResConfig       []byte
		argocdApp              []byte
		repoCreds              []byte
		applyManifests         []byte
		bootstrapKustomization []byte
		namespace              []byte
		externalSecret         []byte
		hashiCorpVault         []byte
		externalSecretManifest []byte
		hashiCorpVaultManifest []byte
	}

	deleteClusterResourcesOptions struct {
		Timeout     time.Duration
		KubeFactory kube.Factory
		FastExit    bool
	}
)

func RunRepoBootstrap(ctx context.Context, opts *RepoBootstrapOptions) error {
	var err error

	if opts, err = setBootstrapOptsDefaults(*opts); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"repo-url":     opts.CloneOptions.URL(),
		"revision":     opts.CloneOptions.Revision(),
		"namespace":    opts.Namespace,
		"kube-context": opts.KubeContextName,
	}).Debug("starting with options: ")

	manifests, err := buildBootstrapManifests(
		opts.Namespace,
		opts.AppSpecifier,
		opts.CloneOptions,
		opts.ArgoCDLabels,
		opts.BootstrapAppsLabels,
		opts.NamespaceLabels,
	)
	if err != nil {
		return fmt.Errorf("failed to build bootstrap manifests: %w", err)
	}

	// Dry Run check
	if opts.DryRun {
		fmt.Printf("%s", util.JoinManifests(
			manifests.namespace,
			manifests.applyManifests,
			manifests.repoCreds,
			manifests.bootstrapApp,
			manifests.argocdApp,
			manifests.rootApp,
		))
		exit(0)
		return nil
	}

	log.Infof("cloning repo: %s", opts.CloneOptions.URL())

	// clone GitOps repo
	r, repofs, err := getRepo(ctx, opts.CloneOptions)
	if err != nil {
		return err
	}

	log.Infof("using revision: \"%s\", installation path: \"%s\"", opts.CloneOptions.Revision(), opts.CloneOptions.Path())
	err = validateRepo(repofs, opts.Recover)
	if err != nil {
		return err
	}

	log.Debug("repository is ok")

	// apply built manifest to k8s cluster
	log.Infof("using context: \"%s\", namespace: \"%s\"", opts.KubeContextName, opts.Namespace)
	log.Infof("applying bootstrap manifests to cluster...")
	if err = opts.KubeFactory.Apply(ctx, util.JoinManifests(manifests.namespace, manifests.applyManifests, manifests.repoCreds)); err != nil {
		return fmt.Errorf("failed to apply bootstrap manifests to cluster: %w", err)
	}

	if !opts.Recover {
		// write argocd manifests to repo
		if err = writeManifestsToRepo(repofs, manifests, opts.InstallationMode, opts.Namespace); err != nil {
			return fmt.Errorf("failed to write manifests to repo: %w", err)
		}
	}

	// wait for argocd to be ready before applying argocd-apps
	stop := util.WithSpinner(ctx, "waiting for argo-cd to be ready")

	if err = waitClusterReady(ctx, opts.KubeFactory, opts.Timeout, opts.Namespace); err != nil {
		stop()
		return err
	}

	stop()

	if !opts.Recover {
		// push results to repo
		log.Infof("pushing bootstrap manifests to repo")
		commitMsg := "feat: supervisor bootstrap"
		if opts.CloneOptions.Path() != "" {
			commitMsg = "supervisor bootstrap at " + opts.CloneOptions.Path()
		}

		if _, err = r.Persist(ctx, &git.PushOptions{CommitMsg: commitMsg}); err != nil {
			return err
		}
	}

	// apply "Argo-CD" Application that references "bootstrap/argo-cd"
	log.Infof("applying argo-cd bootstrap application")
	if err = opts.KubeFactory.Apply(ctx, manifests.bootstrapApp); err != nil {
		return err
	}

	passwd, err := getInitialPassword(ctx, opts.KubeFactory, opts.Namespace)
	if err != nil {
		return err
	}

	if !opts.HidePassword {
		log.Printf("")
		log.Infof("argocd initialized. password: %s", passwd)
		log.Infof("run:\n\n    kubectl port-forward -n %s svc/argocd-server 8080:80\n\n", opts.Namespace)
	}

	return nil
}

func setBootstrapOptsDefaults(opts RepoBootstrapOptions) (*RepoBootstrapOptions, error) {
	var err error
	switch opts.InstallationMode {
	case installationModeFlat, installationModeNormal:
	case "":
		opts.InstallationMode = installationModeNormal
	default:
		return nil, fmt.Errorf("unknown installation mode: %s", opts.InstallationMode)
	}

	if opts.Namespace == "" {
		opts.Namespace = store.Default.ArgoCDNamespace
	}

	if opts.AppSpecifier == "" {
		opts.AppSpecifier = getBootstrapAppSpecifier(opts.Insecure)
	}

	if opts.KubeContextName == "" {
		opts.KubeContextName, err = currentKubeContext()
		if err != nil {
			return &opts, err
		}
	}

	if _, err := os.Stat(opts.AppSpecifier); err == nil {
		log.Warnf("detected local bootstrap manifests, using 'flat' installation mode")
		opts.InstallationMode = installationModeFlat
	}

	return &opts, nil
}

func validateRepo(repofs fs.FS, recover bool) error {
	folders := []string{store.Default.BootsrtrapDir, store.Default.ProjectsDir}
	for _, folder := range folders {
		if repofs.ExistsOrDie(folder) {
			if recover {
				continue
			} else {
				return fmt.Errorf("folder %s already exist in: %s", folder, repofs.Join(repofs.Root(), folder))
			}
		} else if recover {
			return fmt.Errorf("recovery failed: invalid repository, %s directory is missing in %s", folder, repofs.Root())
		}
	}

	return nil
}

func waitClusterReady(ctx context.Context, f kube.Factory, timeout time.Duration, namespace string) error {
	return f.Wait(ctx, &kube.WaitOptions{
		Interval: store.Default.WaitInterval,
		Timeout:  timeout,
		Resources: []kube.Resource{
			{
				Name:      "argocd-server",
				Namespace: namespace,
				WaitFunc:  kube.WaitDeploymentReady,
			},
		},
	})
}

func getRepoCredsSecret(username, token, namespace string) ([]byte, error) {
	return yaml.Marshal(&v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      store.Default.RepoCredsSecretName,
			Namespace: namespace,
			Labels: map[string]string{
				store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
			},
		},
		Data: map[string][]byte{
			"git_username": []byte(username),
			"git_token":    []byte(token),
		},
	})
}

func getInitialPassword(ctx context.Context, f kube.Factory, namespace string) (string, error) {
	cs := f.KubernetesClientSetOrDie()
	secret, err := cs.CoreV1().Secrets(namespace).Get(ctx, "argocd-initial-admin-secret", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	passwd, ok := secret.Data["password"]
	if !ok {
		return "", fmt.Errorf("argocd initial password not found")
	}

	return string(passwd), nil
}

func getBootstrapAppSpecifier(insecure bool) string {
	if insecure {
		return store.Get().InstallationManifestsInsecureURL
	}

	return store.Get().InstallationManifestsURL
}

func buildBootstrapManifests(namespace, appSpecifier string, cloneOpts *git.CloneOptions, argocdLabels map[string]string, bootstrapAppsLabels map[string]string, namespaceLabels map[string]string) (*bootstrapManifests, error) {
	var err error
	manifests := &bootstrapManifests{}

	// h4-bootstrap app
	manifests.bootstrapApp, err = createApp(&createAppOptions{
		name:      store.Default.BootsrtrapAppName,
		namespace: namespace,
		repoURL:   cloneOpts.URL(),
		revision:  cloneOpts.Revision(),
		srcPath:   path.Join(cloneOpts.Path(), store.Default.BootsrtrapDir),
		labels:    bootstrapAppsLabels,
	})
	if err != nil {
		return nil, err
	}

	// root app
	manifests.rootApp, err = createApp(&createAppOptions{
		name:      store.Default.RootAppName,
		namespace: namespace,
		repoURL:   cloneOpts.URL(),
		revision:  cloneOpts.Revision(),
		srcPath:   path.Join(cloneOpts.Path(), store.Default.ProjectsDir),
		labels:    bootstrapAppsLabels,
	})
	if err != nil {
		return nil, err
	}

	// argocd app
	manifests.argocdApp, err = createApp(&createAppOptions{
		name:        store.Default.ArgoCDName,
		namespace:   namespace,
		repoURL:     cloneOpts.URL(),
		revision:    cloneOpts.Revision(),
		srcPath:     path.Join(cloneOpts.Path(), store.Default.BootsrtrapDir, store.Default.ArgoCDName),
		noFinalizer: true,
		labels:      argocdLabels,
	})
	if err != nil {
		return nil, err
	}

	// read external-secrets manifest
	externalSecretManifest, err := os.ReadFile("manifests/external-secrets/external-secrets-manifests.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read external secrets manifest: %w", err)
	}
	manifests.externalSecretManifest = externalSecretManifest

	// read vault manifest
	hashiCorpVaultManifest, err := os.ReadFile("manifests/vault/vault-manifests.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read vault manifest: %w", err)
	}
	manifests.hashiCorpVaultManifest = hashiCorpVaultManifest

	// create external-secrets application
	manifests.externalSecret, err = createApp(&createAppOptions{
		name:        store.Default.BuildInExternalSecrets,
		namespace:   namespace,
		repoURL:     cloneOpts.URL(),
		revision:    cloneOpts.Revision(),
		srcPath:     path.Join(cloneOpts.Path(), store.Default.BootsrtrapDir, store.Default.BuildInExternalSecrets),
		noFinalizer: true,
		labels:      argocdLabels,
	})
	if err != nil {
		return nil, err
	}

	// create vault application
	manifests.hashiCorpVault, err = createApp(&createAppOptions{
		name:        store.Default.BuildInHashiCorpVault,
		namespace:   namespace,
		repoURL:     cloneOpts.URL(),
		revision:    cloneOpts.Revision(),
		srcPath:     path.Join(cloneOpts.Path(), store.Default.BootsrtrapDir, store.Default.BuildInHashiCorpVault),
		noFinalizer: true,
		labels:      argocdLabels,
	})
	if err != nil {
		return nil, err
	}

	// cluster-resources applicationSet
	manifests.clusterResAppSet, err = createAppSet(&createAppSetOptions{
		name:                        store.Default.ClusterResourcesDir,
		namespace:                   namespace,
		repoURL:                     cloneOpts.URL(),
		revision:                    cloneOpts.Revision(),
		appName:                     store.Default.ClusterResourcesDir + "-{{name}}",
		appNamespace:                namespace,
		appLabels:                   bootstrapAppsLabels,
		destServer:                  "{{server}}",
		prune:                       false,
		preserveResourcesOnDeletion: true,
		srcPath:                     path.Join(cloneOpts.Path(), store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, "{{name}}"),
		generators: []argocdv1alpha1.ApplicationSetGenerator{
			{
				Git: &argocdv1alpha1.GitGenerator{
					RepoURL:  cloneOpts.URL(),
					Revision: cloneOpts.Revision(),
					Files: []argocdv1alpha1.GitFileGeneratorItem{
						{
							Path: path.Join(cloneOpts.Path(), store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, "*.json"),
						},
					},
					RequeueAfterSeconds: &DefaultApplicationSetGeneratorInterval,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	manifests.clusterResConfig, err = json.Marshal(&application.ClusterResConfig{Name: store.Default.ClusterContextName, Server: store.Default.DestServer})
	if err != nil {
		return nil, err
	}

	k, err := createBootstrapKustomization(namespace, appSpecifier, cloneOpts)
	if err != nil {
		return nil, err
	}

	if namespace != "" && namespace != "default" {
		ns := kube.GenerateNamespace(namespace, namespaceLabels)
		manifests.namespace, err = yaml.Marshal(ns)
		if err != nil {
			return nil, err
		}
	}

	manifests.applyManifests, err = runKustomizeBuild(k)
	if err != nil {
		return nil, err
	}

	manifests.repoCreds, err = getRepoCredsSecret(cloneOpts.Auth.Username, cloneOpts.Auth.Password, namespace)
	if err != nil {
		return nil, err
	}

	manifests.bootstrapKustomization, err = yaml.Marshal(k)
	if err != nil {
		return nil, err
	}

	return manifests, nil
}

func writeManifestsToRepo(repoFS fs.FS, manifests *bootstrapManifests, installationMode, namespace string) error {
	var bulkWrites []fs.BulkWriteRequest
	argocdPath := repoFS.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName)
	clusterResReadme := []byte(strings.ReplaceAll(string(clusterResReadmeTpl), "{CLUSTER}", store.Default.ClusterContextName))

	if installationMode == installationModeNormal {
		bulkWrites = []fs.BulkWriteRequest{
			{Filename: repoFS.Join(argocdPath, "kustomization.yaml"), Data: manifests.bootstrapKustomization},
		}
	} else {
		bulkWrites = []fs.BulkWriteRequest{
			{Filename: repoFS.Join(argocdPath, "install.yaml"), Data: manifests.applyManifests},
		}
	}

	bulkWrites = append(bulkWrites, []fs.BulkWriteRequest{
		{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.RootAppName+".yaml"), Data: manifests.rootApp},                                                    // write projects root app
		{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName+".yaml"), Data: manifests.argocdApp},                                                   // write argocd app
		{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.BuildInExternalSecrets+".yaml"), Data: manifests.externalSecret},                                  // write external-secrets app
		{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.BuildInHashiCorpVault+".yaml"), Data: manifests.hashiCorpVault},                                   // write vault app
		{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir+".yaml"), Data: manifests.clusterResAppSet},                                   // write cluster-resources appset
		{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, store.Default.ClusterContextName, "README.md"), Data: clusterResReadme},      // write ./bootstrap/cluster-resources/in-cluster/README.md
		{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, store.Default.ClusterContextName+".json"), Data: manifests.clusterResConfig}, // write ./bootstrap/cluster-resources/in-cluster.json
		{Filename: repoFS.Join(store.Default.ProjectsDir, "README.md"), Data: projectReadme},                                                                                // write ./projects/README.md
		{Filename: repoFS.Join(store.Default.AppsDir, "README.md"), Data: appsReadme},                                                                                       // write ./apps/README.md
		{
			Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.BuildInExternalSecrets, "manifests.yaml"),
			Data:     manifests.externalSecretManifest,
		},
		{
			Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.BuildInHashiCorpVault, "manifests.yaml"),
			Data:     manifests.hashiCorpVaultManifest,
		},
	}...)

	if manifests.namespace != nil {
		// write ./bootstrap/cluster-resources/in-cluster/...-ns.yaml
		bulkWrites = append(
			bulkWrites,
			fs.BulkWriteRequest{Filename: repoFS.Join(store.Default.BootsrtrapDir, store.Default.ClusterResourcesDir, store.Default.ClusterContextName, namespace+"-ns.yaml"), Data: manifests.namespace},
		)
	}

	return fs.BulkWrite(repoFS, bulkWrites...)
}

func createBootstrapKustomization(namespace, appSpecifier string, cloneOpts *git.CloneOptions) (*kusttypes.Kustomization, error) {
	credsYAML, err := createCreds(cloneOpts.URL())
	if err != nil {
		return nil, err
	}

	k := &kusttypes.Kustomization{
		Resources: []string{
			appSpecifier,
		},
		TypeMeta: kusttypes.TypeMeta{
			APIVersion: kusttypes.KustomizationVersion,
			Kind:       kusttypes.KustomizationKind,
		},
		ConfigMapGenerator: []kusttypes.ConfigMapArgs{
			{
				GeneratorArgs: kusttypes.GeneratorArgs{
					Name:     "argocd-cm",
					Behavior: kusttypes.BehaviorMerge.String(),
					KvPairSources: kusttypes.KvPairSources{
						LiteralSources: []string{
							"repository.credentials=" + string(credsYAML),
						},
					},
				},
			},
		},
		Namespace: namespace,
	}

	cert, err := cloneOpts.Auth.GetCertificate()
	if err != nil {
		return nil, err
	}

	if cert != nil {
		u, err := url.Parse(cloneOpts.URL())
		if err != nil {
			return nil, err
		}

		host := u.Host
		if strings.Contains(host, ":") {
			host, _, err = net.SplitHostPort(host)
			if err != nil {
				return nil, err
			}
		}

		k.ConfigMapGenerator = append(k.ConfigMapGenerator, kusttypes.ConfigMapArgs{
			GeneratorArgs: kusttypes.GeneratorArgs{
				Name:     "argocd-tls-certs-cm",
				Behavior: kusttypes.BehaviorMerge.String(),
				KvPairSources: kusttypes.KvPairSources{
					LiteralSources: []string{
						host + "=" + string(cert),
					},
				},
			},
		})
	}

	errs := k.EnforceFields()
	if len(errs) > 0 {
		return nil, fmt.Errorf("kustomization errors: %s", strings.Join(errs, "\n"))
	}

	return k, k.FixKustomizationPreMarshalling(filesys.MakeFsInMemory())
}

func createCreds(repoUrl string) ([]byte, error) {
	host, _, _, _, _, _, _ := util.ParseGitUrl(repoUrl)
	creds := []argocdsettings.RepositoryCredentials{
		{
			URL: host,
			UsernameSecret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: store.Default.RepoCredsSecretName,
				},
				Key: "git_username",
			},
			PasswordSecret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: store.Default.RepoCredsSecretName,
				},
				Key: "git_token",
			},
		},
	}

	return yaml.Marshal(creds)
}
