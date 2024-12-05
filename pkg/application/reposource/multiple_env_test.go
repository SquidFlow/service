package reposource

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"

	"github.com/squidflow/service/pkg/fs"
)

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
bases:
- ../../base
`), 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/staging/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
`), 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/prod/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
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
		"mixed valid and invalid overlays": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create overlays with and without kustomization files
				_ = memFS.MkdirAll("overlays/valid", 0666)
				_ = memFS.MkdirAll("overlays/invalid", 0666)

				_ = billyUtils.WriteFile(memFS, "overlays/valid/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
`), 0666)
				return fs.Create(memFS)
			},
			path:     "/",
			wantEnvs: []string{"valid"},
		},
		"custom path": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("custom/path/overlays/dev", 0666)
				_ = billyUtils.WriteFile(memFS, "custom/path/overlays/dev/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
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