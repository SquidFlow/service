package handler

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/h4-poc/service/pkg/util"
	"github.com/slok/go-helm-template/helm"

	"sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func HelmTemplating(app string, env string) error {

	ctx := context.Background()

	chartFS := os.DirFS(fmt.Sprintf("/tmp/platform/manifest/%s/", app))

	chart, err := helm.LoadChart(ctx, chartFS)
	if err != nil {
		return err
	}
	currentMap, err := util.Yaml2Maps(fmt.Sprintf("/tmp/platform/overlays/app/%s/%s/values.yaml", app, env))

	result, err := helm.Template(ctx, helm.TemplateConfig{
		Chart:       chart,
		ReleaseName: app,
		Values: map[string]interface{}{
			"commonLabels": map[string]string{
				"app": app,
			},
			"controller": currentMap,
		},
	})
	if err != nil {
		return err
	}
	err = util.WriteFile(result, fmt.Sprintf("/tmp/platform/overlays/app/%s/%s/manifest.yaml", app, env))
	if err != nil {
		return err
	}
	return nil
}

func KustomizeBuildInOverlay(app string, env string) error {
	fSys := filesys.MakeFsOnDisk()
	buffy := new(bytes.Buffer)
	cmd := build.NewCmdBuild(fSys, build.MakeHelp("foo", "bar"), buffy)
	if err := cmd.RunE(cmd, []string{fmt.Sprintf("/tmp/platform/overlays/app/%s/%s", app, env)}); err != nil {
		return err
	}
	err := util.WriteFile(buffy.String(), fmt.Sprintf("/tmp/platform/overlays/app/%s/%s/generate-manifest.yaml", app, env))
	if err != nil {
		return err
	}
	return nil

}

func KustomizeBuildInManifest(app string, env string) error {
	fSys := filesys.MakeFsOnDisk()
	buffy := new(bytes.Buffer)
	cmd := build.NewCmdBuild(fSys, build.MakeHelp("foo", "bar"), buffy)
	if err := cmd.RunE(cmd, []string{fmt.Sprintf("/tmp/platform/manifest/%s", app)}); err != nil {
		return err
	}
	err := util.WriteFile(buffy.String(), fmt.Sprintf("/tmp/platform/overlays/app/%s/%s/manifest.yaml", app, env))
	if err != nil {
		return err
	}
	return nil

}
