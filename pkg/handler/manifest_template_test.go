package handler

import (
	"bytes"
	"fmt"
	"github.com/squidflow/service/pkg/util"
	"os"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"strings"
	"testing"
)

func KustomizeBuildTest(app string, env string) error {
	fSys := filesys.MakeFsOnDisk()
	buffy := new(bytes.Buffer)
	cmd := build.NewCmdBuild(fSys, build.MakeHelp("foo", "bar"), buffy)
	if err := cmd.RunE(cmd, []string{fmt.Sprintf("/tmp/platform/overlays/app/%s/%s", env, app)}); err != nil {
		fmt.Println("gg")
		fmt.Println(err)
		return err
	}
	err := util.WriteFile(buffy.String(), fmt.Sprintf("/tmp/platform/overlays/app/%s/%s/generate-manifest.yaml", env, app))
	if err != nil {
		fmt.Println("tt")
		fmt.Println(err)
		return err
	}
	return nil

}

func Test_Kustomize(t *testing.T) {
	KustomizeBuildTest("fluent-operator", "sit")
}

func Test_Kubeconform(t *testing.T) {
	data, err := KubeManifestValidator("/tmp/platform/overlays/app/sit/fluent-operator/generate-manifest.yaml")
	fmt.Println(data)
	fmt.Println(err)
}

func Test_Remove(t *testing.T) {
	err := os.RemoveAll("/tmp/platform")
	fmt.Println(err)
}

func Test_Path(t *testing.T) {
	path := "manifest/fluent-operator"
	strList := strings.Split(path, "/")
	app := strList[len(strList)-1]
	overlayPath := fmt.Sprintf("/tmp/platform/overlays/app/%s", app)
	entries, err := os.ReadDir(overlayPath)
	env := map[string]bool{
		"sit":  true,
		"sit1": true,
		"sit2": true,
		"uat":  true,
		"uat1": true,
		"uat2": true,
	}
	if err != nil {
		fmt.Println(err)
	}
	for _, e := range entries {
		if e.Type().IsDir() {
			if env[e.Name()] {
				fmt.Println(e.Name())
			}
		}

	}

}
