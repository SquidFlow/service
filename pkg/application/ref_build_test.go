package application

import "testing"

func TestBuildKustomizeResourceRef(t *testing.T) {
	tests := []struct {
		name string
		args ApplicationSourceOption
		want string
	}{
		{
			name: "simple kustomize ref",
			args: ApplicationSourceOption{
				Repo:           "https://github.com/argoproj/argocd-example-apps.git",
				Path:           "kustomize-guestbook",
				TargetRevision: "master",
			},
			want: "github.com/argoproj/argocd-example-apps/kustomize-guestbook?ref=master",
		},
		{
			name: "git@ format",
			args: ApplicationSourceOption{
				Repo:           "git@github.com:argoproj/argocd-example-apps.git",
				Path:           "kustomize-guestbook",
				TargetRevision: "master",
			},
			want: "github.com/argoproj/argocd-example-apps/kustomize-guestbook?ref=master",
		},
		{
			name: "no target revision",
			args: ApplicationSourceOption{
				Repo: "https://github.com/argoproj/argocd-example-apps.git",
				Path: "kustomize-guestbook",
			},
			want: "github.com/argoproj/argocd-example-apps/kustomize-guestbook?ref=main",
		},
		{
			name: "no path",
			args: ApplicationSourceOption{
				Repo: "https://github.com/argoproj/argocd-example-apps.git",
			},
			want: "github.com/argoproj/argocd-example-apps?ref=main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildKustomizeResourceRef(tt.args); got != tt.want {
				t.Errorf("buildKustomizeResourceRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
