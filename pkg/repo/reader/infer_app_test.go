package reader

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/types"
)

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
