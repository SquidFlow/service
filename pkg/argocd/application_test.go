package argocd

import (
	"reflect"
	"testing"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"gopkg.in/yaml.v2"
)

func TestCreateApp(t *testing.T) {
	tests := []struct {
		name    string
		opts    *CreateAppOptions
		want    string
		wantErr bool
	}{
		{
			name: "Create basic application",
			opts: &CreateAppOptions{
				name:       "test-app",
				namespace:  "argocd",
				repoURL:    "https://github.com/argoproj/argocd-example-apps.git",
				revision:   "HEAD",
				srcPath:    "guestbook",
				destServer: "https://kubernetes.default.svc",
			},
			want: `apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  creationTimestamp: null
  finalizers:
  - resources-finalizer.argocd.argoproj.io
  labels:
    app.kubernetes.io/managed-by: argocd-autopilot
    app.kubernetes.io/name: test-app
  name: test-app
  namespace: argocd
spec:
  destination:
    namespace: argocd
    server: https://kubernetes.default.svc
  ignoreDifferences:
  - group: argoproj.io
    jsonPointers:
    - /status
    kind: Application
  project: default
  source:
    path: guestbook
    repoURL: https://github.com/argoproj/argocd-example-apps.git
    targetRevision: HEAD
  syncPolicy:
    automated:
      allowEmpty: true
      prune: true
      selfHeal: true
    syncOptions:
    - allowEmpty=true
status:
  health: {}
  summary: {}
  sync:
    comparedTo:
      destination: {}
      source:
        repoURL: ""
    status: ""`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateApp(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Unmarshal the 'got' YAML into an Application struct
			var gotApp argocdv1alpha1.Application
			if err := yaml.Unmarshal(got, &gotApp); err != nil {
				t.Errorf("Failed to unmarshal 'got' YAML: %v", err)
				return
			}

			// Unmarshal the 'want' YAML into an Application struct
			var wantApp argocdv1alpha1.Application
			if err := yaml.Unmarshal([]byte(tt.want), &wantApp); err != nil {
				t.Errorf("Failed to unmarshal 'want' YAML: %v", err)
				return
			}

			// Compare the unmarshaled structs
			if !reflect.DeepEqual(gotApp, wantApp) {
				t.Errorf("CreateApp() got = %+v, want %+v", gotApp, wantApp)
			}
		})
	}
}
