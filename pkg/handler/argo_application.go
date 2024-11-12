package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
)

var setAppOptsDefaults = func(ctx context.Context, repofs fs.FS, opts *AppCreateOptions) error {
	var err error

	if opts.createOpts.DestServer == store.Default.DestServer || opts.createOpts.DestServer == "" {
		opts.createOpts.DestServer, err = getProjectDestServer(repofs, opts.ProjectName)
		if err != nil {
			return err
		}
	}

	if opts.createOpts.DestNamespace == "" {
		opts.createOpts.DestNamespace = "default"
	}

	if opts.createOpts.Labels == nil {
		opts.createOpts.Labels = opts.Labels
	}

	if opts.createOpts.Annotations == nil {
		opts.createOpts.Annotations = opts.Annotations
	}

	if opts.createOpts.AppType != "" {
		return nil
	}

	var fsys fs.FS
	if _, err := os.Stat(opts.createOpts.AppSpecifier); err == nil {
		// local directory
		fsys = fs.Create(osfs.New(opts.createOpts.AppSpecifier))
	} else {
		host, orgRepo, p, _, _, suffix, _ := util.ParseGitUrl(opts.createOpts.AppSpecifier)
		url := host + orgRepo + suffix
		log.G().Infof("cloning repo: '%s', to infer app type from path '%s'", url, p)
		cloneOpts := &git.CloneOptions{
			Repo:     opts.createOpts.AppSpecifier,
			Auth:     opts.CloneOpts.Auth,
			Provider: opts.CloneOpts.Provider,
			FS:       fs.Create(memfs.New()),
		}
		cloneOpts.Parse()
		_, fsys, err = getRepo(ctx, cloneOpts)
		if err != nil {
			return err
		}
	}

	opts.createOpts.AppType = application.InferAppType(fsys)
	log.G().Infof("inferred application type: %s", opts.createOpts.AppType)

	return nil
}

var parseApp = func(appOpts *application.CreateOptions, projectName, repoURL, targetRevision, repoRoot string) (application.Application, error) {
	return appOpts.Parse(projectName, repoURL, targetRevision, repoRoot)
}

func getProjectDestServer(repofs fs.FS, projectName string) (string, error) {
	path := repofs.Join(store.Default.ProjectsDir, projectName+".yaml")
	p := &argocdv1alpha1.AppProject{}
	if err := repofs.ReadYamls(path, p); err != nil {
		return "", fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return p.Annotations[store.Default.DestServerAnnotation], nil
}

func genCommitMsg(action ActionType, targetResource ResourceName, appName, projectName string, repofs fs.FS) string {
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
