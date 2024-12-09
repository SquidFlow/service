package dryrun

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/types"
)

func TestGenerateKustomizeManifest(t *testing.T) {
	tests := map[string]struct {
		setupFS   func() fs.FS
		req       types.ApplicationSourceRequest
		env       string
		wantErr   bool
		errString string
	}{
		"simple kustomize": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create a basic kustomization
				_ = billyUtils.WriteFile(memFS, "kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
- service.yaml
`), 0666)
				_ = billyUtils.WriteFile(memFS, "deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test
        image: nginx:latest
`), 0666)
				_ = billyUtils.WriteFile(memFS, "service.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  ports:
  - port: 80
  selector:
    app: test
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:     "default",
			wantErr: false,
		},
		"multi-env kustomize": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create base kustomization
				_ = memFS.MkdirAll("base", 0666)
				_ = billyUtils.WriteFile(memFS, "base/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
`), 0666)
				_ = billyUtils.WriteFile(memFS, "base/deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 1
`), 0666)

				// Create staging overlay
				_ = memFS.MkdirAll("overlays/staging", 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/staging/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
namePrefix: staging-
`), 0666)

				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:     "staging",
			wantErr: false,
		},
		"missing kustomization.yaml": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:       "default",
			wantErr:   true,
			errString: "kustomization.yaml not found in /",
		},
		"invalid overlay": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("overlays/invalid", 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:       "invalid",
			wantErr:   true,
			errString: "kustomization.yaml not found in overlay invalid",
		},
		"multi-env kustomize with patches": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create base kustomization
				_ = memFS.MkdirAll("base", 0666)
				_ = billyUtils.WriteFile(memFS, "base/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
- service.yaml
`), 0666)
				_ = billyUtils.WriteFile(memFS, "base/deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 1
`), 0666)
				_ = billyUtils.WriteFile(memFS, "base/service.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  ports:
  - port: 80
`), 0666)

				// Create prod overlay with patches
				_ = memFS.MkdirAll("overlays/prod", 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/prod/kustomization.yaml", []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
patches:
- path: replica-patch.yaml
`), 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/prod/replica-patch.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 3
`), 0666)

				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:     "prod",
			wantErr: false,
		},
		"overlay with invalid kustomization": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("overlays/invalid", 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/invalid/kustomization.yaml", []byte(`
invalid yaml content
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:       "invalid",
			wantErr:   true,
			errString: "failed to build kustomize",
		},
		"overlay with missing base": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("overlays/dev", 0666)
				_ = billyUtils.WriteFile(memFS, "overlays/dev/kustomization.yaml", []byte(`
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
			env:       "dev",
			wantErr:   true,
			errString: "failed to build kustomize",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()

			result, err := GenerateKustomizeManifest(repofs, tt.req, tt.env, "test-app", "default")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}
