package dryrun

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"

	"github.com/squidflow/service/pkg/fs"
)

func TestGenerateHelmManifest(t *testing.T) {
	tests := map[string]struct {
		setupFS   func() fs.FS
		req       *SourceOption
		env       string
		wantErr   bool
		errString string
	}{
		"simple helm chart": {
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
  tag: latest
service:
  port: 80
`), 0666)
				_ = memFS.MkdirAll("templates", 0666)
				_ = billyUtils.WriteFile(memFS, "templates/deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
`), 0666)
				_ = billyUtils.WriteFile(memFS, "templates/service.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}
spec:
  ports:
  - port: {{ .Values.service.port }}
  selector:
    app: {{ .Release.Name }}
`), 0666)
				return fs.Create(memFS)
			},
			req: &SourceOption{
				Path: "/",
			},
			env:     "default",
			wantErr: false,
		},
		"multi-env helm chart": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create base chart structure
				_ = billyUtils.WriteFile(memFS, "Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)
				_ = billyUtils.WriteFile(memFS, "values.yaml", []byte(`
image:
  repository: nginx
  tag: latest
`), 0666)
				_ = memFS.MkdirAll("templates", 0666)
				_ = billyUtils.WriteFile(memFS, "templates/deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
`), 0666)

				// Create environment-specific values
				_ = memFS.MkdirAll("environments/staging", 0666)
				_ = billyUtils.WriteFile(memFS, "environments/staging/values.yaml", []byte(`
image:
  tag: staging
replicaCount: 2
`), 0666)

				return fs.Create(memFS)
			},
			req: &SourceOption{
				Path: "/",
			},
			env:     "staging",
			wantErr: false,
		},
		"missing Chart.yaml": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			req: &SourceOption{
				Path: "/",
			},
			env:       "default",
			wantErr:   true,
			errString: "Chart.yaml not found at path: /",
		},
		"with specified helm manifest path": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Create a complete chart structure in custom path
				_ = memFS.MkdirAll("custom/path/templates", 0666)

				// Add Chart.yaml
				_ = billyUtils.WriteFile(memFS, "custom/path/Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)

				// Add values.yaml
				_ = billyUtils.WriteFile(memFS, "custom/path/values.yaml", []byte(`
image:
  repository: nginx
  tag: latest
service:
  port: 80
`), 0666)

				// Add template files
				_ = billyUtils.WriteFile(memFS, "custom/path/templates/deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
`), 0666)

				return fs.Create(memFS)
			},
			req: &SourceOption{
				Path: "/",
				SourceSpecifier: &SourceSpecifier{
					HelmManifestPath: "custom/path",
				},
			},
			env:     "default",
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()

			result, err := GenerateHelmManifest(repofs, tt.req, tt.env, "test-app", "default")

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