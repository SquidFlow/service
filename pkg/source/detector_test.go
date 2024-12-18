package source

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
		wantSourceType string
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
			wantSourceType: AppTypeHelm,
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
			wantSourceType: AppTypeKustomize,
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
			wantSourceType: AppTypeHelmMultiEnv,
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
			wantSourceType: AppTypeKustomizeMultiEnv,
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
				ApplicationSpecifier: types.ApplicationSpecifier{
					HelmManifestPath: "custom/path",
				},
			},
			wantSourceType: AppTypeHelmMultiEnv,
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
		"kustomize with base and overlays but no kustomization files": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("base", 0666)
				_ = memFS.MkdirAll("overlays/dev", 0666)
				_ = memFS.MkdirAll("overlays/prod", 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeKustomizeMultiEnv,
			wantEnvs:       []string{"default"},
			wantErr:        false,
		},
		"kustomize with kustomization.yml": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = billyUtils.WriteFile(memFS, "kustomization.yml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeKustomize,
			wantEnvs:       []string{"default"},
			wantErr:        false,
		},
		"helm with environments but no values files": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("manifests", 0666)
				_ = billyUtils.WriteFile(memFS, "manifests/Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
`), 0666)
				_ = memFS.MkdirAll("environments/dev", 0666)
				_ = memFS.MkdirAll("environments/prod", 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeHelmMultiEnv,
			wantEnvs:       []string{"default"},
			wantErr:        false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()
			sourceType, envs, err := InferApplicationSource(repofs, tt.req)

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

func TestDetectApplicationType(t *testing.T) {
	tests := map[string]struct {
		setupFS        func() fs.FS
		req            types.ApplicationSourceRequest
		wantSourceType string
		wantErr        bool
	}{
		"helm with application specifier": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
				ApplicationSpecifier: types.ApplicationSpecifier{
					HelmManifestPath: "manifests/helm",
				},
			},
			wantSourceType: AppTypeHelmMultiEnv,
			wantErr:        false,
		},
		"kustomize with kustomization.yaml": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = billyUtils.WriteFile(memFS, "kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeKustomize,
			wantErr:        false,
		},
		"helm with Chart.yaml": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = billyUtils.WriteFile(memFS, "Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeHelm,
			wantErr:        false,
		},
		"helm multi-env with manifests and environments": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("manifests", 0666)
				_ = memFS.MkdirAll("environments", 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeHelmMultiEnv,
			wantErr:        false,
		},
		"kustomize multi-env with base and overlays": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("base", 0666)
				_ = memFS.MkdirAll("overlays", 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeKustomizeMultiEnv,
			wantErr:        false,
		},
		"unsupported structure, default is dir": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("random", 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			wantSourceType: AppTypeDirectory,
			wantErr:        false,
		},
		"nested path with kustomization": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("nested/path", 0666)
				_ = billyUtils.WriteFile(memFS, "nested/path/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "nested/path",
			},
			wantSourceType: AppTypeKustomize,
			wantErr:        false,
		},
		"nested path with Chart.yaml": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("nested/path", 0666)
				_ = billyUtils.WriteFile(memFS, "nested/path/Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "nested/path",
			},
			wantSourceType: AppTypeHelm,
			wantErr:        false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()
			sourceType, err := detectApplicationType(repofs, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSourceType, sourceType)
		})
	}
}

func TestInferAppType(t *testing.T) {
	tests := map[string]struct {
		want     string
		beforeFn func() fs.FS
	}{
		"Should return ksonnet if required files are present": {
			want: "ksonnet",
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "app.yaml", []byte{}, 0666)
				_ = billyUtils.WriteFile(memfs, "components/params.libsonnet", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should not return ksonnet if 'app.yaml' is missing": {
			want: AppTypeDirectory,
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "components/params.libsonnet", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should not return ksonnet if 'components/params.libsonnet' is missing": {
			want: AppTypeDirectory,
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "app.yaml", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should return ksonnet as the highest priority": {
			want: "ksonnet",
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "app.yaml", []byte{}, 0666)
				_ = billyUtils.WriteFile(memfs, "components/params.libsonnet", []byte{}, 0666)
				_ = billyUtils.WriteFile(memfs, "Chart.yaml", []byte{}, 0666)
				_ = billyUtils.WriteFile(memfs, "kustomization.yaml", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should return helm if 'Chart.yaml' is present": {
			want: "helm",
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "Chart.yaml", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should return kustomize as a higher priority than helm": {
			want: "kustomize",
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "Chart.yaml", []byte{}, 0666)
				_ = billyUtils.WriteFile(memfs, "kustomization.yaml", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should return kustomize if 'kustomization.yaml' file is present": {
			want: AppTypeKustomize,
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "kustomization.yaml", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should return kustomize if 'kustomization.yml' file is present": {
			want: AppTypeKustomize,
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "kustomization.yml", []byte{}, 0666)
				return fs.Create(memfs)
			},
		},
		"Should return kustomize if 'Kustomization' folder is present": {
			want: AppTypeKustomize,
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = memfs.MkdirAll("Kustomization", 0666)
				return fs.Create(memfs)
			},
		},
		"Should return dir if no other match": {
			want: AppTypeDirectory,
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				return fs.Create(memfs)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.beforeFn()
			if got := InferAppType(repofs); got != tt.want {
				t.Errorf("InferAppType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelmMultiEnvDetector_DetectEnvironments(t *testing.T) {
	tests := map[string]struct {
		setupFS   func() fs.FS
		path      string
		wantEnvs  []string
		wantError bool
	}{
		"standard environments directory": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create environments with values files
				_ = memFS.MkdirAll("environments/dev", 0666)
				_ = memFS.MkdirAll("environments/staging", 0666)
				_ = memFS.MkdirAll("environments/prod", 0666)

				_ = billyUtils.WriteFile(memFS, "environments/dev/values.yaml", []byte(`
environment: dev
`), 0666)
				_ = billyUtils.WriteFile(memFS, "environments/staging/values.yaml", []byte(`
environment: staging
`), 0666)
				_ = billyUtils.WriteFile(memFS, "environments/prod/values.yaml", []byte(`
environment: prod
`), 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"dev", "staging", "prod"},
		},
		"empty environments directory": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("environments", 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"default"},
		},
		"no environments directory": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			path:     "/",
			wantEnvs: []string{"default"},
		},
		"mixed valid and invalid environments": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create environments with and without values files
				_ = memFS.MkdirAll("environments/valid", 0666)
				_ = memFS.MkdirAll("environments/invalid", 0666)

				_ = billyUtils.WriteFile(memFS, "environments/valid/values.yaml", []byte(`
environment: valid
`), 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"valid"},
		},
		"custom path": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("custom/path/environments/dev", 0666)
				_ = billyUtils.WriteFile(memFS, "custom/path/environments/dev/values.yaml", []byte(`
environment: dev
`), 0666)
				return fs.Create(memFS)
			},
			path:     "custom/path",
			wantEnvs: []string{"dev"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()
			detector := &HelmMultiEnvDetector{
				repofs: repofs,
				path:   tt.path,
			}

			envs, err := detector.DetectEnvironments()
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.wantEnvs, envs)
		})
	}
}

func TestKustomizeMultiEnvDetector_DetectEnvironments(t *testing.T) {
	tests := map[string]struct {
		setupFS   func() fs.FS
		path      string
		wantEnvs  []string
		wantError bool
	}{
		"standard overlays directory": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create overlays with kustomization files
				_ = memFS.MkdirAll("overlays/dev", 0666)
				_ = memFS.MkdirAll("overlays/staging", 0666)
				_ = memFS.MkdirAll("overlays/prod", 0666)

				_ = billyUtils.WriteFile(memFS, "overlays/dev/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/staging/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/prod/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"dev", "staging", "prod"},
		},
		"empty overlays directory": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("overlays", 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"default"},
		},
		"no overlays directory": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			path:     "/",
			wantEnvs: []string{"default"},
		},
		"overlays with kustomization.yml": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("overlays/dev", 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/dev/kustomization.yml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"dev"},
		},
		"overlays with both yaml and yml": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("overlays/dev", 0666)
				_ = memFS.MkdirAll("overlays/prod", 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/dev/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/prod/kustomization.yml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`), 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"dev", "prod"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()
			detector := &KustomizeMultiEnvDetector{
				repofs: repofs,
				path:   tt.path,
			}

			envs, err := detector.DetectEnvironments()
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.wantEnvs, envs)
		})
	}
}
