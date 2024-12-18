package dryrun

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/types"
)

func TestGenerateHelmManifest(t *testing.T) {
	tests := map[string]struct {
		setupFS   func() fs.FS
		req       types.ApplicationSourceRequest
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
			req: types.ApplicationSourceRequest{
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
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:     "staging",
			wantErr: false,
		},
		"missing Chart.yaml": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			req: types.ApplicationSourceRequest{
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
			req: types.ApplicationSourceRequest{
				Path: "/",
				ApplicationSpecifier: types.ApplicationSpecifier{
					HelmManifestPath: "custom/path",
				},
			},
			env:     "default",
			wantErr: false,
		},
		"helm chart with environment values override": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Base chart
				_ = billyUtils.WriteFile(memFS, "Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)
				// Default values
				_ = billyUtils.WriteFile(memFS, "values.yaml", []byte(`
image:
  repository: nginx
  tag: latest
replicaCount: 1
`), 0666)
				// Environment specific values
				_ = memFS.MkdirAll("environments/prod", 0666)
				_ = billyUtils.WriteFile(memFS, "environments/prod/values.yaml", []byte(`
image:
  tag: stable
replicaCount: 3
`), 0666)
				// Templates
				_ = memFS.MkdirAll("templates", 0666)
				_ = billyUtils.WriteFile(memFS, "templates/deployment.yaml", []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  template:
    spec:
      containers:
      - image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:     "prod",
			wantErr: false,
		},
		"missing environment values file": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = billyUtils.WriteFile(memFS, "Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)
				_ = memFS.MkdirAll("environments/missing", 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:       "missing",
			wantErr:   true,
			errString: "failed to read values file",
		},
		"invalid values.yaml syntax": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = billyUtils.WriteFile(memFS, "Chart.yaml", []byte(`
apiVersion: v2
name: test-chart
version: 0.1.0
`), 0666)
				_ = billyUtils.WriteFile(memFS, "values.yaml", []byte(`
invalid: yaml: content:
  - missing colon
  unclosed quote"
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:       "default",
			wantErr:   true,
			errString: "failed to parse values.yaml",
		},
		"invalid template syntax": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
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
metadata:
  name: {{ .Release.Name }
`), 0666)
				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
			},
			env:       "default",
			wantErr:   true,
			errString: "failed to render templates",
		},
		"custom helm manifest path with subcharts": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				// Main chart
				_ = memFS.MkdirAll("charts/main/charts/subchart", 0666)
				// Main chart files
				_ = billyUtils.WriteFile(memFS, "charts/main/Chart.yaml", []byte(`
apiVersion: v2
name: main-chart
version: 0.1.0
dependencies:
- name: subchart
  version: 0.1.0
`), 0666)
				// add main chart values.yaml
				_ = billyUtils.WriteFile(memFS, "charts/main/values.yaml", []byte(`
subchart:
  enabled: true
  config:
    key: value
`), 0666)
				// add main chart templates
				_ = memFS.MkdirAll("charts/main/templates", 0666)
				_ = billyUtils.WriteFile(memFS, "charts/main/templates/configmap.yaml", []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
data:
  key: {{ .Values.subchart.config.key }}
`), 0666)

				// Subchart files
				_ = billyUtils.WriteFile(memFS, "charts/main/charts/subchart/Chart.yaml", []byte(`
apiVersion: v2
name: subchart
version: 0.1.0
`), 0666)
				// add subchart values.yaml
				_ = billyUtils.WriteFile(memFS, "charts/main/charts/subchart/values.yaml", []byte(`
config:
  key: default-value
`), 0666)
				// add subchart templates
				_ = memFS.MkdirAll("charts/main/charts/subchart/templates", 0666)
				_ = billyUtils.WriteFile(memFS, "charts/main/charts/subchart/templates/service.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-subchart
spec:
  ports:
  - port: 80
`), 0666)

				return fs.Create(memFS)
			},
			req: types.ApplicationSourceRequest{
				Path: "/",
				ApplicationSpecifier: types.ApplicationSpecifier{
					HelmManifestPath: "charts/main",
				},
			},
			env:     "default",
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()
			var path, manifestPath string
			if tt.req.ApplicationSpecifier.HelmManifestPath != "" {
				path = tt.req.Path
				manifestPath = tt.req.ApplicationSpecifier.HelmManifestPath
			} else {
				path = tt.req.Path
				manifestPath = ""
			}

			result, err := GenerateHelmManifest(repofs, path, manifestPath, tt.env, "test-app", "default")

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
