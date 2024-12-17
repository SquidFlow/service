package writer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	reporeader "github.com/squidflow/service/pkg/repo/reader"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"
	"github.com/squidflow/service/pkg/util"

	appmocks "github.com/squidflow/service/pkg/application/mocks"
	fsmocks "github.com/squidflow/service/pkg/fs/mocks"
	gitmocks "github.com/squidflow/service/pkg/git/mocks"
	kubemocks "github.com/squidflow/service/pkg/kube/mocks"
)

func TestRunAppCreate(t *testing.T) {
	tests := map[string]struct {
		repoWriter            NativeRepoTarget
		wantErr               string
		setAppOptsDefaultsErr error
		parseAppErr           error
		createFilesErr        error
		beforeFn              func(f *kubemocks.MockFactory)
		prepareRepo           func(*testing.T) (git.Repository, fs.FS, error)
		getRepo               func(*testing.T, *git.CloneOptions) (git.Repository, fs.FS, error)
	}{
		"Should fail when clone fails": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr: "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should fail if srcClone fails": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr: "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
			getRepo: func(_ *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should use cloneOpts password for srcCloneOpts, if required": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/other_name",
					Auth: git.Auth{
						Password: "password",
					},
				},
				tenantRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/other_name/path?ref=branch",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr: "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, nil
			},
			getRepo: func(t *testing.T, opts *git.CloneOptions) (git.Repository, fs.FS, error) {
				assert.Equal(t, "https://github.com/owner/other_name.git", opts.URL())
				assert.Equal(t, "branch", opts.Revision())
				assert.Equal(t, "path", opts.Path())
				assert.Equal(t, "password", opts.Auth.Password)
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should fail if setAppOptsDefaults fails": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.xxxxx/xxxxx",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr: "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, nil
			},
			setAppOptsDefaultsErr: fmt.Errorf("some error"),
		},
		"Should fail if app parse fails": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.xxxxx/xxxxx",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr: "failed to parse application from flags: some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, nil
			},
			parseAppErr: errors.New("some error"),
		},
		"Should fail if app already exist in project": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr:        fmt.Errorf("application 'app' already exists in project 'project': %w", application.ErrAppAlreadyInstalledOnProject).Error(),
			createFilesErr: application.ErrAppAlreadyInstalledOnProject,
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				return mockRepo, fs.Create(memfs), nil
			},
		},
		"Should fail if file creation fails": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr:        "some error",
			createFilesErr: errors.New("some error"),
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				return mockRepo, fs.Create(memfs), nil
			},
		},
		"Should fail if committing to appsRepo fails": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
				tenantRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/other_name",
				},
			},
			wantErr: "failed to push to apps repo: some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				return mockRepo, fs.Create(memfs), nil
			},
			getRepo: func(_ *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: create app 'app' on project 'project' installation-path: '/'",
				}).
					Times(1).
					Return("", fmt.Errorf("some error"))
				return mockRepo, fs.Create(memfs.New()), nil
			},
		},
		"Should fail if committing to gitops repo fails": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			wantErr: "failed to push to gitops repo: some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: create app 'app' on project 'project' installation-path: '/'",
				}).
					Times(1).
					Return("", fmt.Errorf("some error"))
				return mockRepo, fs.Create(memfs), nil
			},
		},
		// TODO: add send commit to tenant repo
		//"Should Persist to both repos, if required": {
		//	repoWriter: NativeRepoTarget{
		//		metaRepoCloneOpts: &git.CloneOptions{
		//			Repo: "https://github.com/owner/name",
		//			Auth: git.Auth{
		//				Password: "password",
		//			},
		//		},
		//		tenantRepoCloneOpts: &git.CloneOptions{
		//			Repo: "https://github.com/owner/other_name",
		//		},
		//	},
		//	prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
		//		memfs := memfs.New()
		//		_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
		//		mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
		//		mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
		//			CommitMsg: "chore: create app 'app' on project 'project' installation-path: '/'",
		//		}).
		//			Times(1).
		//			Return("revision", nil)
		//		return mockRepo, fs.Create(memfs), nil
		//	},
		//	getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
		//		mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
		//		mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
		//			CommitMsg: "chore: create app 'app' on project 'project' installation-path: '/'",
		//		}).
		//			Times(1).
		//			Return("revision", nil)
		//		return mockRepo, fs.Create(memfs.New()), nil
		//	},
		//},
		"Should Persist to a single repo, if required": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: create app 'app' on project 'project' installation-path: '/'",
				}).
					Times(1).
					Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
		},
	}
	origPrepareRepo, origGetRepo, origSetAppOptsDefault, origAppParse := prepareRepo, getRepo, setAppOptsDefaults, parseApp
	defer func() {
		prepareRepo = origPrepareRepo
		getRepo = origGetRepo
		setAppOptsDefaults = origSetAppOptsDefault
		parseApp = origAppParse
	}()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				gitopsRepo git.Repository
				appsRepo   git.Repository
			)

			ctrl := gomock.NewController(t)
			f := kubemocks.NewMockFactory(ctrl)
			if tt.beforeFn != nil {
				tt.beforeFn(f)
			}
			prepareRepo = func(_ context.Context, _ *git.CloneOptions, _ string) (git.Repository, fs.FS, error) {
				var (
					repofs fs.FS
					err    error
				)
				gitopsRepo, repofs, err = tt.prepareRepo(t)
				return gitopsRepo, repofs, err
			}
			getRepo = func(_ context.Context, cloneOpts *git.CloneOptions) (git.Repository, fs.FS, error) {
				var (
					repofs fs.FS
					err    error
				)
				appsRepo, repofs, err = tt.getRepo(t, cloneOpts)
				return appsRepo, repofs, err
			}
			setAppOptsDefaults = func(_ context.Context, _ fs.FS, _ *application.AppCreateOptions) error {
				return tt.setAppOptsDefaultsErr
			}
			parseApp = func(_ *application.CreateOptions, _, _, _, _ string) (application.Application, error) {
				if tt.parseAppErr != nil {
					return nil, tt.parseAppErr
				}

				app := appmocks.NewMockApplication(ctrl)
				app.EXPECT().Name().Return("app").AnyTimes()
				app.EXPECT().CreateFiles(gomock.Any(), gomock.Any(), "project").Return(tt.createFilesErr).AnyTimes()
				return app, nil
			}
			opts := &application.AppCreateOptions{
				ProjectName: "project",
				AppOpts: &application.CreateOptions{
					AppName:      "app",
					AppType:      reporeader.AppTypeDirectory,
					AppSpecifier: "https://github.com/owner/name/manifests",
				},
				KubeFactory: f,
			}

			// meta repo
			tt.repoWriter.metaRepoCloneOpts.Parse()
			if tt.repoWriter.tenantRepoCloneOpts == nil {
				tt.repoWriter.tenantRepoCloneOpts = tt.repoWriter.metaRepoCloneOpts
			}
			tt.repoWriter.tenantRepoCloneOpts.Parse()

			// create app
			if err := tt.repoWriter.RunAppCreate(context.Background(), opts); err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestRunAppDelete(t *testing.T) {
	tests := map[string]struct {
		repoWriter  NativeRepoTarget
		appName     string
		projectName string
		global      bool
		wantErr     string
		getRepo     func(*testing.T, *git.CloneOptions) (git.Repository, fs.FS, error)
		assertFn    func(t *testing.T, repo git.Repository, repofs fs.FS)
	}{
		"Should fail when clone fails": {
			repoWriter: NativeRepoTarget{
				project: "project",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
						Username: "username",
					},
				},
			},
			wantErr: "some error",
			getRepo: func(_ *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should fail when app does not exist": {
			repoWriter: NativeRepoTarget{
				project: "project",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			wantErr: "application 'app' not found",
			getRepo: func(*testing.T, *git.CloneOptions) (git.Repository, fs.FS, error) {
				return nil, fs.Create(memfs.New()), nil
			},
		},
		"Should fail if deletion of entire app directory fails": {
			repoWriter: NativeRepoTarget{
				project: "",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			global:  true,
			wantErr: fmt.Sprintf("failed to delete directory '%s': some error", filepath.Join(store.Default.AppsDir, "app")),
			getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				mfs := fsmocks.NewMockFS(gomock.NewController(t))
				path := filepath.Join(store.Default.AppsDir, "app")
				mfs.EXPECT().Join(gomock.Any()).
					Times(1).
					DoAndReturn(func(elem ...string) string {
						return strings.Join(elem, "/")
					})
				mfs.EXPECT().ExistsOrDie(path).Return(true)
				mfs.EXPECT().Remove(path).Return(fmt.Errorf("some error"))
				mfs.EXPECT().Stat(path).Return(nil, fmt.Errorf("some error"))
				return nil, mfs, nil
			},
		},
		"Should remove entire app directory when global flag is set": {
			repoWriter: NativeRepoTarget{
				project: "",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			global:  true,
			getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: delete app 'app'",
				}).
					Times(1).
					Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")))
			},
		},
		"Should fail when overlay does not exist": {
			repoWriter: NativeRepoTarget{
				project: "project",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			wantErr: "application 'app' not found in project 'project'",
			getRepo: func(*testing.T, *git.CloneOptions) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project2"), 0666)
				return nil, fs.Create(memfs), nil
			},
		},
		"Should fail if ReadDir fails": {
			repoWriter: NativeRepoTarget{
				project: "project",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			wantErr: fmt.Sprintf("failed to read overlays directory '%s': some error", filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)),
			getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				mfs := fsmocks.NewMockFS(gomock.NewController(t))
				mfs.EXPECT().Join(gomock.Any()).
					Times(3).
					DoAndReturn(func(elem ...string) string {
						return strings.Join(elem, "/")
					})
				mfs.EXPECT().ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")).
					Times(1).
					Return(true)
				mfs.EXPECT().ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)).
					Times(1).
					Return(true)
				mfs.EXPECT().ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project")).
					Times(1).
					Return(true)
				mfs.EXPECT().ReadDir(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)).
					Times(1).
					Return(nil, fmt.Errorf("some error"))
				return nil, mfs, nil
			},
		},
		"Should delete only overlay directory of a kustApp, if there are more overlays": {
			repoWriter: NativeRepoTarget{
				project: "project",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project2"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: delete app 'app' from project 'project'",
				}).
					Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project")))
			},
		},
		"Should delete entire app directory of a kustApp, if there are no more overlays": {
			repoWriter: NativeRepoTarget{
				project: "project",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: delete app 'app'",
				}).
					Times(1).
					Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")))
			},
		},
		"Should fail if Persist fails": {
			repoWriter: NativeRepoTarget{
				project: "",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			appName: "app",
			wantErr: "failed to push to repo: some error",
			getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: delete app 'app'",
				}).
					Times(1).
					Return("", fmt.Errorf("some error"))
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")))
			},
		},
	}
	origGetRepo := getRepo
	defer func() {
		getRepo = origGetRepo
	}()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				repo   git.Repository
				repofs fs.FS
			)

			getRepo = func(_ context.Context, cloneOpts *git.CloneOptions) (git.Repository, fs.FS, error) {
				var err error
				repo, repofs, err = tt.getRepo(t, cloneOpts)
				return repo, repofs, err
			}

			tt.repoWriter.metaRepoCloneOpts.Parse()
			tt.repoWriter.tenantRepoCloneOpts = tt.repoWriter.metaRepoCloneOpts
			if err := tt.repoWriter.RunAppDelete(context.Background(), tt.appName); err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if tt.assertFn != nil {
				tt.assertFn(t, repo, repofs)
			}
		})
	}
}

func TestRunProjectCreate(t *testing.T) {
	tests := map[string]struct {
		repoWriter               NativeRepoTarget
		projectName              string
		wantErr                  string
		getInstallationNamespace func(repofs fs.FS) (string, error)
		prepareRepo              func(*testing.T) (git.Repository, fs.FS, error)
		assertFn                 func(t *testing.T, repo git.Repository, repofs fs.FS)
	}{
		"should handle failure in prepare repo": {
			repoWriter:  NativeRepoTarget{},
			projectName: "project",
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("failure clone")
			},
			wantErr: "failure clone",
		},
		"should handle failure while getting namespace": {
			repoWriter:  NativeRepoTarget{},
			projectName: "project",
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				return nil, fs.Create(memfs), nil
			},
			getInstallationNamespace: func(_ fs.FS) (string, error) {
				return "", fmt.Errorf("failure namespace")
			},
			wantErr: util.Doc("Bootstrap folder not found, please execute `<BIN> repo bootstrap --installation-path /` command"),
		},
		"should handle failure when project exists": {
			repoWriter:  NativeRepoTarget{},
			projectName: "project",
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = billyUtils.WriteFile(memfs, "projects/project.yaml", []byte{}, 0666)
				return nil, fs.Create(memfs), nil
			},
			getInstallationNamespace: func(_ fs.FS) (string, error) {
				return "namespace", nil
			},
			wantErr: "project 'project' already exists",
		},
		"should handle failure when writing project file": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "http://server.org/org/name",
				},
			},
			projectName: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				mockedFS := fsmocks.NewMockFS(gomock.NewController(t))
				mockedFS.EXPECT().Join("projects", "project.yaml").
					Times(2).
					Return("projects/project.yaml")
				mockedFS.EXPECT().ExistsOrDie("projects/project.yaml").Return(false)
				mockedFS.EXPECT().OpenFile("projects/project.yaml", gomock.Any(), gomock.Any()).Return(nil, os.ErrPermission)
				return nil, mockedFS, nil
			},
			getInstallationNamespace: func(_ fs.FS) (string, error) {
				return "namespace", nil
			},
			wantErr: "failed to create project file: permission denied",
		},
		"should handle failure to persist repo": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "http://server.org/org/name",
				},
			},
			projectName: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				mockedRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockedRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: added project 'project'",
				}).Return("", fmt.Errorf("failed to persist"))
				return mockedRepo, fs.Create(memfs), nil
			},
			getInstallationNamespace: func(_ fs.FS) (string, error) {
				return "namespace", nil
			},
			wantErr: "failed to persist",
		},
		"should persist repo when done": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "http://server.org/org/name",
				},
			},
			projectName: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				mockedRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockedRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: added project 'project'",
				}).Return("revision", nil)
				return mockedRepo, fs.Create(memfs), nil
			},
			getInstallationNamespace: func(_ fs.FS) (string, error) {
				return "namespace", nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				exists := repofs.ExistsOrDie("projects/project.yaml")
				assert.True(t, exists)
			},
		},
	}
	origPrepareRepo := prepareRepo
	origGetInstallationNamespace := getInstallationNamespace
	defer func() {
		prepareRepo = origPrepareRepo
		getInstallationNamespace = origGetInstallationNamespace
	}()

	for ttName, tt := range tests {
		t.Run(ttName, func(t *testing.T) {
			var (
				repo   git.Repository
				repofs fs.FS
			)

			opts := &types.ProjectCreateOptions{
				ProjectName: tt.projectName,
			}
			prepareRepo = func(_ context.Context, _ *git.CloneOptions, _ string) (git.Repository, fs.FS, error) {
				var err error
				repo, repofs, err = tt.prepareRepo(t)
				return repo, repofs, err
			}
			getInstallationNamespace = tt.getInstallationNamespace
			if err := tt.repoWriter.RunProjectCreate(context.Background(), opts); err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if tt.assertFn != nil {
				tt.assertFn(t, repo, repofs)
			}
		})
	}
}

func Test_generateProjectManifests(t *testing.T) {
	tests := map[string]struct {
		o                      *types.GenerateProjectOptions
		wantName               string
		wantNamespace          string
		wantProjectDescription string
		wantRepoURL            string
		wantRevision           string
		wantDefaultDestServer  string
		wantProject            string
		wantContextName        string
		wantLabels             map[string]string
		wantAnnotations        map[string]string
	}{
		"should generate project and appset with correct values": {
			o: &types.GenerateProjectOptions{
				Name:               "name",
				Namespace:          "namespace",
				DefaultDestServer:  "defaultDestServer",
				DefaultDestContext: "some-context-name",
				RepoURL:            "repoUrl",
				Revision:           "revision",
				InstallationPath:   "some/path",
				Labels: map[string]string{
					"some-key": "some-value",
				},
				Annotations: map[string]string{
					"some-key": "some-value",
				},
			},
			wantName:               "name",
			wantNamespace:          "namespace",
			wantProjectDescription: "name project",
			wantRepoURL:            "repoUrl",
			wantRevision:           "revision",
			wantDefaultDestServer:  "defaultDestServer",
			wantContextName:        "some-context-name",
			wantLabels: map[string]string{
				"some-key":                         "some-value",
				store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
				store.Default.LabelKeyAppName:      "{{ appName }}",
			},
			wantAnnotations: map[string]string{
				"some-key": "some-value",
			},
		},
		"set tenant repo url": {
			o: &types.GenerateProjectOptions{
				Name:               "name",
				Namespace:          "namespace",
				DefaultDestServer:  "defaultDestServer",
				DefaultDestContext: "some-context-name",
				ProjectGitopsRepo:  "tenantRepoUrl",
				RepoURL:            "repoUrl",
				Revision:           "revision",
				InstallationPath:   "some/path",
				Labels: map[string]string{
					"some-key": "some-value",
				},
				Annotations: map[string]string{
					"some-key": "some-value",
				},
			},
			wantName:               "name",
			wantNamespace:          "namespace",
			wantProjectDescription: "name project",
			wantRepoURL:            "tenantRepoUrl",
			wantRevision:           "revision",
			wantDefaultDestServer:  "defaultDestServer",
			wantContextName:        "some-context-name",
			wantLabels: map[string]string{
				"some-key":                         "some-value",
				store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
				store.Default.LabelKeyAppName:      "{{ appName }}",
			},
			wantAnnotations: map[string]string{
				"some-key": "some-value",
			},
		},
	}
	for ttname, tt := range tests {
		t.Run(ttname, func(t *testing.T) {
			assert := assert.New(t)
			gotProject := &argocdv1alpha1.AppProject{}
			gotAppSet := &argocdv1alpha1.ApplicationSet{}
			gotClusterResConf := &application.ClusterResConfig{}
			gotProjectYAML, gotAppSetYAML, _, gotClusterResConfigYAML, _ := generateProjectManifests(tt.o)
			assert.NoError(yaml.Unmarshal(gotProjectYAML, gotProject))
			assert.NoError(yaml.Unmarshal(gotAppSetYAML, gotAppSet))
			assert.NoError(yaml.Unmarshal(gotClusterResConfigYAML, gotClusterResConf))

			assert.Equal(tt.wantContextName, gotClusterResConf.Name)
			assert.Equal(tt.wantDefaultDestServer, gotClusterResConf.Server)

			assert.Equal(tt.wantName, gotProject.Name, "Project Name")
			assert.Equal(tt.wantNamespace, gotProject.Namespace, "Project Namespace")
			assert.Equal(tt.wantProjectDescription, gotProject.Spec.Description, "Project Description")
			assert.Equal(tt.o.DefaultDestServer, gotProject.Annotations[store.Default.DestServerAnnotation], "Application Set Default Destination Server")

			assert.Equal(tt.wantName, gotAppSet.Name, "Application Set Name")
			assert.Equal(tt.wantNamespace, gotAppSet.Namespace, "Application Set Namespace")
			assert.Equal(tt.wantRepoURL, gotAppSet.Spec.Generators[0].Git.RepoURL, "Application Set Repo URL")
			assert.Equal(tt.wantRevision, gotAppSet.Spec.Generators[0].Git.Revision, "Application Set Revision")

			assert.Equal(tt.wantNamespace, gotAppSet.Spec.Template.Namespace, "Application Set Template Namespace")
			assert.Equal(tt.wantName, gotAppSet.Spec.Template.Spec.Project, "Application Set Template Project")
		})
	}
}

func Test_getInstallationNamespace(t *testing.T) {
	tests := map[string]struct {
		beforeFn func(*testing.T) fs.FS
		want     string
		wantErr  string
	}{
		"should return the namespace from namespace.yaml": {
			beforeFn: func(*testing.T) fs.FS {
				namespace := &argocdv1alpha1.Application{
					Spec: argocdv1alpha1.ApplicationSpec{
						Destination: argocdv1alpha1.ApplicationDestination{
							Namespace: "namespace",
						},
					},
				}
				repofs := fs.Create(memfs.New())
				_ = repofs.WriteYamls(filepath.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName+".yaml"), namespace)
				return repofs
			},
			want: "namespace",
		},
		"should handle file not found": {
			beforeFn: func(*testing.T) fs.FS {
				return fs.Create(memfs.New())
			},
			wantErr: "failed to unmarshal namespace: file does not exist",
		},
		"should handle error during read": {
			beforeFn: func(t *testing.T) fs.FS {
				mfs := fsmocks.NewMockFS(gomock.NewController(t))
				mfs.EXPECT().Join(gomock.Any()).
					Times(1).
					DoAndReturn(func(elem ...string) string {
						return strings.Join(elem, "/")
					})
				mfs.EXPECT().ReadYamls(filepath.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName+".yaml"), gomock.Any()).
					Times(1).
					Return(fmt.Errorf("some error"))
				return mfs
			},
			wantErr: "failed to unmarshal namespace: some error",
		},
		"should handle corrupted namespace.yaml file": {
			beforeFn: func(*testing.T) fs.FS {
				repofs := fs.Create(memfs.New())
				_ = billyUtils.WriteFile(repofs, filepath.Join(store.Default.BootsrtrapDir, store.Default.ArgoCDName+".yaml"), []byte("some string"), 0666)
				return repofs
			},
			wantErr: "failed to unmarshal namespace: error unmarshaling JSON: json: cannot unmarshal string into Go value of type v1alpha1.Application",
		},
	}
	for ttName, tt := range tests {
		t.Run(ttName, func(t *testing.T) {
			repofs := tt.beforeFn(t)
			got, err := getInstallationNamespace(repofs)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("getInstallationNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getProjectInfoFromFile(t *testing.T) {
	tests := map[string]struct {
		name     string
		want     *argocdv1alpha1.AppProject
		wantErr  string
		beforeFn func(repofs fs.FS)
	}{
		"should return error if project file doesn't exist": {
			name:    "prod.yaml",
			wantErr: os.ErrNotExist.Error(),
		},
		"should failed when 2 files not found": {
			name:    "prod.yaml",
			wantErr: "expected at least 2 manifests when reading 'prod.yaml'",
			beforeFn: func(repofs fs.FS) {
				_ = billyUtils.WriteFile(repofs, "prod.yaml", []byte("content"), 0666)
			},
		},
		"should return AppProject": {
			name: "prod.yaml",
			beforeFn: func(repofs fs.FS) {
				appProj := argocdv1alpha1.AppProject{
					ObjectMeta: v1.ObjectMeta{
						Name:      "prod",
						Namespace: "ns",
					},
				}
				appSet := argocdv1alpha1.ApplicationSet{}
				_ = repofs.WriteYamls("prod.yaml", appProj, appSet)
			},
			want: &argocdv1alpha1.AppProject{
				ObjectMeta: v1.ObjectMeta{
					Name:      "prod",
					Namespace: "ns",
				},
			},
		},
	}
	for tName, tt := range tests {
		t.Run(tName, func(t *testing.T) {
			repofs := fs.Create(memfs.New())
			if tt.beforeFn != nil {
				tt.beforeFn(repofs)
			}

			got, _, err := getProjectInfoFromFile(repofs, tt.name)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getProjectInfoFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunProjectList(t *testing.T) {
	tests := map[string]struct {
		repoWriter  NativeRepoTarget
		prepareRepo func(*testing.T) (git.Repository, fs.FS, error)
		wantErr     string
		assertFn    func(t *testing.T, projects []types.TenantInfo)
	}{
		"should handle failure in prepare repo": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "http://server.org/org/name",
				},
			},
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("failure clone")
			},
			wantErr: "failure clone",
		},
		"should list projects": {
			repoWriter: NativeRepoTarget{
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "http://server.org/org/name",
				},
			},
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				appProj := argocdv1alpha1.AppProject{
					ObjectMeta: v1.ObjectMeta{
						Name:      "prod",
						Namespace: "ns",
						Annotations: map[string]string{
							store.Default.DestServerAnnotation: "https://cluster.example.com",
						},
					},
				}
				appSet := argocdv1alpha1.ApplicationSet{
					ObjectMeta: v1.ObjectMeta{
						Name: "prod",
					},
					Spec: argocdv1alpha1.ApplicationSetSpec{
						Generators: []argocdv1alpha1.ApplicationSetGenerator{
							{
								Git: &argocdv1alpha1.GitGenerator{
									RepoURL: "https://example.org/org/name.git",
								},
							},
						},
					},
				}
				repofs := fs.Create(memfs)
				_ = repofs.WriteYamls("projects/prod.yaml", appProj, appSet)
				return nil, repofs, nil
			},
			assertFn: func(t *testing.T, projects []types.TenantInfo) {
				assert.Len(t, projects, 1)
				assert.Equal(t, "prod", projects[0].Name)
				assert.Equal(t, "ns", projects[0].Namespace)
				assert.Equal(t, "https://cluster.example.com", projects[0].DefaultCluster)
			},
		},
	}

	origPrepareRepo := prepareRepo
	defer func() { prepareRepo = origPrepareRepo }()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				repo   git.Repository
				repofs fs.FS
			)

			prepareRepo = func(_ context.Context, _ *git.CloneOptions, _ string) (git.Repository, fs.FS, error) {
				var err error
				repo, repofs, err = tt.prepareRepo(t)
				return repo, repofs, err
			}

			projects, err := tt.repoWriter.RunProjectList(context.Background())
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if tt.assertFn != nil {
				tt.assertFn(t, projects)
			}
		})
	}
}

func TestRunProjectDelete(t *testing.T) {
	tests := map[string]struct {
		repoWriter            NativeRepoTarget
		projectNameWantDelete string
		wantErr               string
		prepareRepo           func(*testing.T) (git.Repository, fs.FS, error)
		assertFn              func(t *testing.T, repo git.Repository, repofs fs.FS)
	}{
		"Should fail when clone fails": {
			repoWriter: NativeRepoTarget{
				project: "admin",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			projectNameWantDelete: "project",
			wantErr:               "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should fail when failed to delete project.yaml file": {
			repoWriter: NativeRepoTarget{
				project: "admin",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			projectNameWantDelete: "project",
			wantErr:               "failed to delete project 'project': " + os.ErrNotExist.Error(),
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: deleted project 'project'",
				}).Times(0)
				return mockRepo, fs.Create(memfs), nil
			},
		},
		"Should fail when persist fails": {
			repoWriter: NativeRepoTarget{
				project: "admin",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			projectNameWantDelete: "project",
			wantErr:               "failed to push to repo: some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project"), 0666)
				_ = billyUtils.WriteFile(memfs, filepath.Join(store.Default.ProjectsDir, "project.yaml"), []byte{}, 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: deleted project 'project'",
				}).Return("", fmt.Errorf("some error"))
				return mockRepo, fs.Create(memfs), nil
			},
		},
		"Should remove entire app folder, if it contains only one overlay": {
			repoWriter: NativeRepoTarget{
				project: "admin",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			projectNameWantDelete: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project"), 0666)
				_ = billyUtils.WriteFile(memfs, filepath.Join(store.Default.ProjectsDir, "project.yaml"), []byte{}, 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: deleted project 'project'",
				}).Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app1")))
			},
		},
		"Should remove only overlay, if app contains more overlays": {
			repoWriter: NativeRepoTarget{
				project: "admin",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			projectNameWantDelete: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project2"), 0666)
				_ = billyUtils.WriteFile(memfs, filepath.Join(store.Default.ProjectsDir, "project.yaml"), []byte{}, 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: deleted project 'project'",
				}).Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir)))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project")))
			},
		},
		"Should remove directory apps": {
			repoWriter: NativeRepoTarget{
				project: "admin",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			projectNameWantDelete: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", "project"), 0666)
				_ = billyUtils.WriteFile(memfs, filepath.Join(store.Default.ProjectsDir, "project.yaml"), []byte{}, 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: deleted project 'project'",
				}).Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app1")))
			},
		},
		"Should handle multiple apps": {
			repoWriter: NativeRepoTarget{
				project: "admin",
				metaRepoCloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
			},
			projectNameWantDelete: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project2"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app2", store.Default.OverlaysDir, "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app3", "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app4", "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app4", "project3"), 0666)
				_ = billyUtils.WriteFile(memfs, filepath.Join(store.Default.ProjectsDir, "project.yaml"), []byte{}, 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(context.Background(), &git.PushOptions{
					CommitMsg: "chore: deleted project 'project'",
				}).Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir)))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app1", store.Default.OverlaysDir, "project")))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app2")))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app3")))
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app4")))
			},
		},
	}
	origPrepareRepo := prepareRepo
	defer func() { prepareRepo = origPrepareRepo }()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				repo   git.Repository
				repofs fs.FS
			)

			prepareRepo = func(_ context.Context, _ *git.CloneOptions, _ string) (git.Repository, fs.FS, error) {
				var err error
				repo, repofs, err = tt.prepareRepo(t)
				return repo, repofs, err
			}
			tt.repoWriter.metaRepoCloneOpts.Parse()

			if err := tt.repoWriter.RunProjectDelete(context.Background(), tt.projectNameWantDelete); err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if tt.assertFn != nil {
				tt.assertFn(t, repo, repofs)
			}
		})
	}
}

func Test_getDefaultAppLabels(t *testing.T) {
	tests := map[string]struct {
		labels map[string]string
		want   map[string]string
	}{
		"Should return the default map when sending nil": {
			labels: nil,
			want: map[string]string{
				store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
				store.Default.LabelKeyAppName:      "{{ appName }}",
			},
		},
		"Should contain any additional labels sent": {
			labels: map[string]string{
				"something": "or the other",
			},
			want: map[string]string{
				"something":                        "or the other",
				store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
				store.Default.LabelKeyAppName:      "{{ appName }}",
			},
		},
		"Should overwrite the default managed by": {
			labels: map[string]string{
				store.Default.LabelKeyAppManagedBy: "someone else",
			},
			want: map[string]string{
				store.Default.LabelKeyAppManagedBy: "someone else",
				store.Default.LabelKeyAppName:      "{{ appName }}",
			},
		},
		"Should overwrite the default app name": {
			labels: map[string]string{
				store.Default.LabelKeyAppName: "another name",
			},
			want: map[string]string{
				store.Default.LabelKeyAppManagedBy: store.Default.LabelValueManagedBy,
				store.Default.LabelKeyAppName:      "another name",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := getDefaultAppLabels(tt.labels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDefaultAppLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
