package repotarget

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/squidflow/service/pkg/application"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/types"

	appmocks "github.com/squidflow/service/pkg/application/mocks"
	fsmocks "github.com/squidflow/service/pkg/fs/mocks"
	gitmocks "github.com/squidflow/service/pkg/git/mocks"
	kubemocks "github.com/squidflow/service/pkg/kube/mocks"
)

func TestRunAppCreate(t *testing.T) {
	native := NativeRepoTarget{}

	tests := map[string]struct {
		appsRepo                 string
		timeout                  time.Duration
		wantErr                  string
		setAppOptsDefaultsErr    error
		parseAppErr              error
		createFilesErr           error
		beforeFn                 func(f *kubemocks.MockFactory)
		prepareRepo              func(*testing.T) (git.Repository, fs.FS, error)
		getRepo                  func(*testing.T, *git.CloneOptions) (git.Repository, fs.FS, error)
		getInstallationNamespace func(repofs fs.FS) (string, error)
	}{
		"Should fail when clone fails": {
			wantErr: "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should fail if srcClone fails": {
			appsRepo: "https://github.com/owner/other_name",
			wantErr:  "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, nil
			},
			getRepo: func(_ *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should use cloneOpts password for srcCloneOpts, if required": {
			appsRepo: "https://github.com/owner/other_name/path?ref=branch",
			wantErr:  "some error",
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
			wantErr: "some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, nil
			},
			setAppOptsDefaultsErr: fmt.Errorf("some error"),
		},
		"Should fail if app parse fails": {
			wantErr: "failed to parse application from flags: some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, nil
			},
			parseAppErr: errors.New("some error"),
		},
		"Should fail if app already exist in project": {
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
			appsRepo: "https://github.com/owner/other_name",
			wantErr:  "failed to push to apps repo: some error",
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
		"Should fail if getInstallationNamespace fails": {
			timeout: 1,
			wantErr: "failed to get application namespace: some error",
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
			getInstallationNamespace: func(repofs fs.FS) (string, error) {
				return "", errors.New("some error")
			},
		},
		"Should fail if waiting fails": {
			timeout: 1,
			wantErr: "failed waiting for application to sync: some error",
			beforeFn: func(f *kubemocks.MockFactory) {
				f.EXPECT().Wait(gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("some error"))
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
			getInstallationNamespace: func(repofs fs.FS) (string, error) {
				return "namespace", nil
			},
		},
		"Should Persist to both repos, if required": {
			appsRepo: "https://github.com/owner/other_name",
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
			getRepo: func(t *testing.T, _ *git.CloneOptions) (git.Repository, fs.FS, error) {
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: create app 'app' on project 'project' installation-path: '/'",
				}).
					Times(1).
					Return("revision", nil)
				return mockRepo, fs.Create(memfs.New()), nil
			},
		},
		"Should Persist to a single repo, if required": {
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
		"Should wait succesfully and complete": {
			timeout: 1,
			beforeFn: func(f *kubemocks.MockFactory) {
				f.EXPECT().Wait(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
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
			getInstallationNamespace: func(repofs fs.FS) (string, error) {
				return "namespace", nil
			},
		},
	}
	origPrepareRepo, origGetRepo, origSetAppOptsDefault, origAppParse, origGetInstallationNamespace := prepareRepo, getRepo, setAppOptsDefaults, parseApp, getInstallationNamespace
	defer func() {
		prepareRepo = origPrepareRepo
		getRepo = origGetRepo
		setAppOptsDefaults = origSetAppOptsDefault
		parseApp = origAppParse
		getInstallationNamespace = origGetInstallationNamespace
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
			setAppOptsDefaults = func(_ context.Context, _ fs.FS, _ *types.AppCreateOptions) error {
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
			getInstallationNamespace = tt.getInstallationNamespace
			opts := &types.AppCreateOptions{
				Timeout: tt.timeout,
				CloneOpts: &git.CloneOptions{
					Repo: "https://github.com/owner/name",
					Auth: git.Auth{
						Password: "password",
					},
				},
				AppsCloneOpts: &git.CloneOptions{
					Repo: tt.appsRepo,
				},
				ProjectName: "project",
				AppOpts: &application.CreateOptions{
					AppName:      "app",
					AppType:      application.AppTypeDirectory,
					AppSpecifier: "https://github.com/owner/name/manifests",
				},
				KubeFactory: f,
			}

			opts.CloneOpts.Parse()
			opts.AppsCloneOpts.Parse()
			if err := native.RunAppCreate(context.Background(), opts); err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestRunAppDelete(t *testing.T) {
	native := NativeRepoTarget{}
	tests := map[string]struct {
		appName     string
		projectName string
		global      bool
		wantErr     string
		prepareRepo func(*testing.T) (git.Repository, fs.FS, error)
		assertFn    func(t *testing.T, repo git.Repository, repofs fs.FS)
	}{
		"Should fail when clone fails": {
			wantErr: "some error",
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				return nil, nil, fmt.Errorf("some error")
			},
		},
		"Should fail when app does not exist": {
			appName: "app",
			wantErr: "application 'app' not found",
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				return nil, fs.Create(memfs.New()), nil
			},
		},
		"Should fail if deletion of entire app directory fails": {
			appName: "app",
			global:  true,
			wantErr: fmt.Sprintf("failed to delete directory '%s': some error", filepath.Join(store.Default.AppsDir, "app")),
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
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
			appName: "app",
			global:  true,
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
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
			appName:     "app",
			projectName: "project",
			wantErr:     "application 'app' not found in project 'project'",
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project2"), 0666)
				return nil, fs.Create(memfs), nil
			},
		},
		"Should fail if ReadDir fails": {
			appName:     "app",
			projectName: "project",
			wantErr:     fmt.Sprintf("failed to read overlays directory '%s': some error", filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)),
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
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
			appName:     "app",
			projectName: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project2"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: delete app 'app' from project 'project'",
				}).
					Times(1).
					Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project")))
			},
		},
		"Should delete entire app directory of a kustApp, if there are no more overlays": {
			appName:     "app",
			projectName: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
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
		"Should delete only project directory of a dirApp, if there are more projects": {
			appName:     "app",
			projectName: "project",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", "project2"), 0666)
				mockRepo := gitmocks.NewMockRepository(gomock.NewController(t))
				mockRepo.EXPECT().Persist(gomock.Any(), &git.PushOptions{
					CommitMsg: "chore: delete app 'app' from project 'project'",
				}).
					Times(1).
					Return("revision", nil)
				return mockRepo, fs.Create(memfs), nil
			},
			assertFn: func(t *testing.T, _ git.Repository, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", "project")))
			},
		},
		"Should delete entire app directory of a dirApp": {
			appName:     "app",
			projectName: "project",
			prepareRepo: func(*testing.T) (git.Repository, fs.FS, error) {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", "project"), 0666)
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
			appName: "app",
			global:  true,
			wantErr: "failed to push to repo: some error",
			prepareRepo: func(t *testing.T) (git.Repository, fs.FS, error) {
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
			opts := &types.AppDeleteOptions{
				ProjectName: tt.projectName,
				AppName:     tt.appName,
				Global:      tt.global,
			}
			if err := native.RunAppDelete(context.Background(), opts); err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if tt.assertFn != nil {
				tt.assertFn(t, repo, repofs)
			}
		})
	}
}
