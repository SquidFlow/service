package repowriter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"k8s.io/client-go/kubernetes"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/argocd"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/kube"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"
	"github.com/squidflow/service/pkg/util"
)

var (
	getRepo = func(ctx context.Context, cloneOpts *git.CloneOptions) (git.Repository, fs.FS, error) {
		return cloneOpts.GetRepo(ctx)
	}

	prepareRepo = func(ctx context.Context, cloneOpts *git.CloneOptions, projectName string) (git.Repository, fs.FS, error) {
		log.G().WithFields(log.Fields{
			"repo-url":      cloneOpts.URL(),
			"repo-revision": cloneOpts.Revision(),
			"repo-path":     cloneOpts.Path(),
		}).Debugf("starting with options:")

		log.G().Infof("cloning git repository: %s", cloneOpts.URL())
		r, repofs, err := getRepo(ctx, cloneOpts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed cloning the repository: %w", err)
		}

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

	setAppOptsDefaults = func(ctx context.Context, repofs fs.FS, opts *types.AppCreateOptions) error {
		var err error

		if opts.AppOpts.DestServer == store.Default.DestServer || opts.AppOpts.DestServer == "" {
			opts.AppOpts.DestServer, err = getProjectDestServer(repofs, opts.ProjectName)
			if err != nil {
				return err
			}
		}

		if opts.AppOpts.DestNamespace == "" {
			opts.AppOpts.DestNamespace = "default"
		}

		if opts.AppOpts.Labels == nil {
			opts.AppOpts.Labels = opts.AppOpts.Labels
		}

		if opts.AppOpts.Annotations == nil {
			opts.AppOpts.Annotations = opts.AppOpts.Annotations
		}

		if opts.AppOpts.AppType != "" {
			return nil
		}

		var fsys fs.FS
		if _, err := os.Stat(opts.AppOpts.AppSpecifier); err == nil {
			// local directory
			fsys = fs.Create(osfs.New(opts.AppOpts.AppSpecifier))
		} else {
			host, orgRepo, p, _, _, suffix, _ := util.ParseGitUrl(opts.AppOpts.AppSpecifier)
			url := host + orgRepo + suffix
			log.G().Infof("cloning repo: '%s', to infer app type from path '%s'", url, p)
			cloneOpts := &git.CloneOptions{
				Repo: opts.AppOpts.AppSpecifier,
				FS:   fs.Create(memfs.New()),
			}
			cloneOpts.Parse()
			_, fsys, err = getRepo(ctx, cloneOpts)
			if err != nil {
				return err
			}
		}

		opts.AppOpts.AppType = application.InferAppType(fsys)
		log.G().Infof("inferred application type: %s", opts.AppOpts.AppType)

		return nil
	}

	parseApp = func(appOpts *application.CreateOptions, projectName, repoURL, targetRevision, repoRoot string) (application.Application, error) {
		return appOpts.Parse(projectName, repoURL, targetRevision, repoRoot)
	}
)

func getProjectDestServer(repofs fs.FS, projectName string) (string, error) {
	path := repofs.Join(store.Default.ProjectsDir, projectName+".yaml")
	p := &argocdv1alpha1.AppProject{}
	if err := repofs.ReadYamls(path, p); err != nil {
		return "", fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return p.Annotations[store.Default.DestServerAnnotation], nil
}

func genCommitMsg(action types.ActionType, targetResource types.ResourceName, appName, projectName string, repofs fs.FS) string {
	commitMsg := fmt.Sprintf("%s %s '%s' on project '%s'", action, targetResource, appName, projectName)
	if repofs.Root() != "" {
		commitMsg += fmt.Sprintf(" installation-path: '%s'", repofs.Root())
	}

	return commitMsg
}

func getConfigFileFromPath(repofs fs.FS, appPath string) (*application.Config, error) {
	path := repofs.Join(appPath, "config.json")
	b, err := repofs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s'", path)
	}

	conf := application.Config{}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file '%s'", path)
	}

	return &conf, nil
}

// TODO: Implement this function later
func getGitInfo(repofs billy.Filesystem, appPath string) (*types.GitInfo, error) {
	return &types.GitInfo{
		Creator:           "Unknown",
		LastUpdater:       "Unknown",
		LastCommitID:      "Unknown",
		LastCommitMessage: "Unknown",
	}, nil
}

// TODO: Implement this function later
func getResourceMetrics(ctx context.Context, kubeClient kubernetes.Interface, namespace string) (*types.ResourceMetricsInfo, error) {
	return &types.ResourceMetricsInfo{
		PodCount:    5,
		SecretCount: 12,
		CPU:         "0.25",
		Memory:      "200Mi",
	}, nil
}

var getInstallationNamespace = func(repofs fs.FS) (string, error) {
	path := repofs.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName+".yaml")
	a := &argocdv1alpha1.Application{}
	if err := repofs.ReadYamls(path, a); err != nil {
		return "", fmt.Errorf("failed to unmarshal namespace: %w", err)
	}

	return a.Spec.Destination.Namespace, nil
}

func waitAppSynced(ctx context.Context, f kube.Factory, timeout time.Duration, appName, namespace, revision string, waitForCreation bool) error {
	return f.Wait(ctx, &kube.WaitOptions{
		Interval: store.Default.WaitInterval,
		Timeout:  timeout,
		Resources: []kube.Resource{
			{
				Name:      appName,
				Namespace: namespace,
				WaitFunc:  argocd.GetAppSyncWaitFunc(revision, waitForCreation),
			},
		},
	})
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
