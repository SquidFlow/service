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
	log "github.com/sirupsen/logrus"

	"github.com/h4-poc/service/pkg/application"
	"github.com/h4-poc/service/pkg/argocd"
	"github.com/h4-poc/service/pkg/fs"
	"github.com/h4-poc/service/pkg/git"
	"github.com/h4-poc/service/pkg/kube"
	"github.com/h4-poc/service/pkg/store"
	"github.com/h4-poc/service/pkg/util"
)

var setAppOptsDefaults = func(ctx context.Context, repofs fs.FS, opts *AppCreateOptions) error {
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
		opts.AppOpts.Labels = opts.Labels
	}

	if opts.AppOpts.Annotations == nil {
		opts.AppOpts.Annotations = opts.Annotations
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
		log.Infof("cloning repo: '%s', to infer app type from path '%s'", url, p)
		cloneOpts := &git.CloneOptions{
			Repo:     opts.AppOpts.AppSpecifier,
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

	opts.AppOpts.AppType = application.InferAppType(fsys)
	log.Infof("inferred application type: %s", opts.AppOpts.AppType)

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

func getCommitMsg(opts *AppCreateOptions, repofs fs.FS) string {
	commitMsg := fmt.Sprintf("installed app '%s' on project '%s'", opts.AppOpts.AppName, opts.ProjectName)
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
