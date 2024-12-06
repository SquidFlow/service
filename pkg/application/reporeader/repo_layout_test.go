package reporeader

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/types"
)

func TestValidateApplicationStructure(t *testing.T) {
	tests := map[string]struct {
		setupFS        func() fs.FS
		req            types.ApplicationSourceRequest
		wantSourceType AppSourceType
		wantEnvs       []string
		wantErr        bool
	}{
		"helm single env": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create a basic Helm chart structure
				_ = billyUtils.WriteFile(memFS, "Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)
				_ = billyUtils.WriteFile(memFS, "values.yaml", []byte(`
image:
  repository: nginx
`), 0666)
				_ = memFS.MkdirAll("templates", 0666)
				_ = billyUtils.WriteFile(memFS, "templates/deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: SourceHelm,
			wantEnvs:       []string{"default"},
			wantErr:        false,
		},
		"kustomize single env": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create a basic Kustomize structure
				_ = billyUtils.WriteFile(memFS, "kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
`), 0666)
				_ = billyUtils.WriteFile(memFS, "deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: SourceKustomizeMultiEnv,
			wantEnvs:       []string{"default"},
			wantErr:        false,
		},
		"helm multi env": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create a Helm chart with multiple environments
				_ = memFS.MkdirAll("manifests", 0666)
				_ = billyUtils.WriteFile(memFS, "manifests/Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)

				// Create environment-specific values
				_ = memFS.MkdirAll("environments/staging", 0666)
				_ = memFS.MkdirAll("environments/production", 0666)
				_ = billyUtils.WriteFile(memFS, "environments/staging/values.yaml", []byte(`
environment: staging
`), 0666)
				_ = billyUtils.WriteFile(memFS, "environments/production/values.yaml", []byte(`
environment: production
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: SourceHelmMultiEnv,
			wantEnvs:       []string{"staging", "production"},
			wantErr:        false,
		},
		"kustomize multi env": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create a Kustomize structure with base and overlays
				_ = memFS.MkdirAll("base", 0666)
				_ = billyUtils.WriteFile(memFS, "base/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
`), 0666)

				// Create environment-specific overlays
				_ = memFS.MkdirAll("overlays/staging", 0666)
				_ = memFS.MkdirAll("overlays/production", 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/staging/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
`), 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/production/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: SourceKustomizeMultiEnv,
			wantEnvs:       []string{"staging", "production"},
			wantErr:        false,
		},
		"helm with specified manifest path": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create a custom Helm chart structure
				_ = memFS.MkdirAll("custom/path", 0666)
				_ = billyUtils.WriteFile(memFS, "custom/path/Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
				ApplicationSpecifier: &types.ApplicationSpecifier{
					HelmManifestPath: "custom/path",
				},
			},
			wantSourceType: SourceHelmMultiEnv,
			wantEnvs:       []string{"default"},
			wantErr:        false,
		},
		"invalid path": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			req: types.ApplicationSourceRequest{
				Path: "/nonexistent",
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()
			sourceType, envs, err := ValidateApplicationStructure(repofs, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSourceType, sourceType)
			assert.ElementsMatch(t, tt.wantEnvs, envs)
		})
	}
}
