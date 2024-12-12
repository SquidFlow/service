package git

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/squidflow/service/pkg/git/bitbucket-server/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	providerOptions = &ProviderOptions{
		Auth: &Auth{
			Username: "username",
			Password: "password",
		},
	}
)

func baseURL() *url.URL {
	u, _ := url.Parse("https://some.server")
	return u
}

func createBody(obj interface{}) io.ReadCloser {
	data, _ := json.Marshal(obj)
	return io.NopCloser(strings.NewReader(string(data)))
}

func Test_bitbucketServer_CreateRepository(t *testing.T) {
	tests := map[string]struct {
		orgRepo  string
		want     string
		wantErr  string
		beforeFn func(t *testing.T, c *mocks.MockHttpClient)
	}{
		"Should fail if orgRepo is invalid": {
			orgRepo: "no-scm/project/repo",
			wantErr: "invalid Bitbucket url \"no-scm/project/repo\" - must be in the form of \"scm/[~]project-or-username/repo-name\"",
		},
		"Should fail if repos POST fails": {
			orgRepo: "scm/project/repo",
			wantErr: "some error",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).Return(nil, errors.New("some error"))
			},
		},
		"Should create a valid project repo": {
			orgRepo: "scm/project/repo",
			want:    "main",
			beforeFn: func(t *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "POST", req.Method)
					assert.Equal(t, "https://some.server/rest/api/1.0/projects/project/repos", req.URL.String())
					repo := &repoResponse{
						DefaultBranch: "main",
					}
					body := createBody(repo)
					res := &http.Response{
						StatusCode: 200,
						Body:       body,
					}
					return res, nil
				})
			},
		},
		"Should create a valid user repo": {
			orgRepo: "scm/~user/repo",
			beforeFn: func(t *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "POST", req.Method)
					assert.Equal(t, "https://some.server/rest/api/1.0/users/user/repos", req.URL.String())
					repo := &repoResponse{
						Links: Links{
							Clone: []Link{
								{
									Name: "https",
									Href: "https://some.server/scm/~user/repo.git",
								},
							},
						},
					}
					body := createBody(repo)
					res := &http.Response{
						StatusCode: 200,
						Body:       body,
					}
					return res, nil
				})
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockClient := mocks.NewMockHttpClient(ctrl)
			if tt.beforeFn != nil {
				tt.beforeFn(t, mockClient)
			}

			bbs := &bitbucketServer{
				baseURL: baseURL(),
				c:       mockClient,
				opts:    providerOptions,
			}
			got, err := bbs.CreateRepository(context.Background(), tt.orgRepo)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "CreateRepository - %s", name)
		})
	}
}

func Test_bitbucketServer_GetDefaultBranch(t *testing.T) {
	tests := map[string]struct {
		orgRepo  string
		want     string
		wantErr  string
		beforeFn func(t *testing.T, c *mocks.MockHttpClient)
	}{
		"Should fail if orgRepo is invalid": {
			orgRepo: "no-scm/project/repo",
			wantErr: "invalid Bitbucket url \"no-scm/project/repo\" - must be in the form of \"scm/[~]project-or-username/repo-name\"",
		},
		"Should fail if repos GET fails": {
			orgRepo: "scm/project/repo",
			wantErr: "some error",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).Return(nil, errors.New("some error"))
			},
		},
		"Should return defaultBranch from project repo": {
			orgRepo: "scm/project/repo",
			want:    "some-branch",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "https://some.server/rest/api/1.0/projects/project/repos/repo", req.URL.String())
					repo := &repoResponse{
						DefaultBranch: "some-branch",
					}
					body := createBody(repo)
					res := &http.Response{
						StatusCode: 200,
						Body:       body,
					}
					return res, nil
				})
			},
		},
		"Should return defaultBranch from user repo": {
			orgRepo: "scm/~user/repo",
			want:    "some-branch",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "https://some.server/rest/api/1.0/users/user/repos/repo", req.URL.String())
					repo := &repoResponse{
						DefaultBranch: "some-branch",
					}
					body := createBody(repo)
					res := &http.Response{
						StatusCode: 200,
						Body:       body,
					}
					return res, nil
				})
			},
		},
		"Should return 'master' if no defaultBranch is set": {
			orgRepo: "scm/project/repo",
			want:    "master",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				repo := &repoResponse{}
				body := createBody(repo)
				res := &http.Response{
					StatusCode: 200,
					Body:       body,
				}
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).Return(res, nil)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockClient := mocks.NewMockHttpClient(ctrl)
			if tt.beforeFn != nil {
				tt.beforeFn(t, mockClient)
			}

			bbs := &bitbucketServer{
				baseURL: baseURL(),
				c:       mockClient,
				opts:    providerOptions,
			}
			got, err := bbs.GetDefaultBranch(context.Background(), tt.orgRepo)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)

		})
	}
}

func Test_bitbucketServer_GetAuthor(t *testing.T) {
	tests := map[string]struct {
		wantUsername string
		wantEmail    string
		wantErr      string
		beforeFn     func(t *testing.T, c *mocks.MockHttpClient)
	}{
		"Should fail if whoami GET fails": {
			wantErr: "failed getting current user's slug: some error",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).Return(nil, errors.New("some error"))
			},
		},
		"Should fail if user GET fails": {
			wantErr: "failed getting current user: some error",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				callFirst := c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "https://some.server/plugins/servlet/applinks/whoami", req.URL.String())
					res := &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(string("username"))),
					}
					return res, nil
				})
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).Return(nil, errors.New("some error")).After(callFirst)
			},
		},
		"Should return displayName and emailAddress if available": {
			wantUsername: "displayName",
			wantEmail:    "username@email",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				callFirst := c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "https://some.server/plugins/servlet/applinks/whoami", req.URL.String())
					res := &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(string("username"))),
					}
					return res, nil
				})
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "https://some.server/rest/api/1.0/users/username", req.URL.String())
					user := &userResponse{
						DisplayName:  "displayName",
						EmailAddress: "username@email",
					}
					res := &http.Response{
						StatusCode: 200,
						Body:       createBody(user),
					}
					return res, nil
				}).After(callFirst)
			},
		},
		"Should return name and slug if no displayName and emailAddress": {
			wantUsername: "name",
			wantEmail:    "slug",
			beforeFn: func(_ *testing.T, c *mocks.MockHttpClient) {
				callFirst := c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "https://some.server/plugins/servlet/applinks/whoami", req.URL.String())
					res := &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(string("username"))),
					}
					return res, nil
				})
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "GET", req.Method)
					assert.Equal(t, "https://some.server/rest/api/1.0/users/username", req.URL.String())
					user := &userResponse{
						Name: "name",
						Slug: "slug",
					}
					res := &http.Response{
						StatusCode: 200,
						Body:       createBody(user),
					}
					return res, nil
				}).After(callFirst)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockClient := mocks.NewMockHttpClient(ctrl)
			if tt.beforeFn != nil {
				tt.beforeFn(t, mockClient)
			}

			bbs := &bitbucketServer{
				baseURL: baseURL(),
				c:       mockClient,
				opts:    providerOptions,
			}
			gotUsername, gotEmail, err := bbs.GetAuthor(context.Background())
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantUsername, gotUsername, "username mismatch")
			assert.Equal(t, tt.wantEmail, gotEmail, "email mismatch")
		})
	}
}

func Test_splitOrgRepo(t *testing.T) {
	tests := map[string]struct {
		orgRepo   string
		wantNoun  string
		wantOwner string
		wantName  string
		wantErr   string
	}{
		"Should fail if it doesn't start with 'scm'": {
			orgRepo: "no-scm-start/project/repo",
			wantErr: "invalid Bitbucket url \"no-scm-start/project/repo\" - must be in the form of \"scm/[~]project-or-username/repo-name\"",
		},
		"Should fail if it contains more segments": {
			orgRepo: "scm/project/sub/repo",
			wantErr: "invalid Bitbucket url \"scm/project/sub/repo\" - must be in the form of \"scm/[~]project-or-username/repo-name\"",
		},
		"Should succeed for a simple orgRepo": {
			orgRepo:   "scm/project/repo",
			wantNoun:  "projects",
			wantOwner: "project",
			wantName:  "repo",
		},
		"Should identify ~ as users": {
			orgRepo:   "scm/~user/repo",
			wantNoun:  "users",
			wantOwner: "user",
			wantName:  "repo",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotNoun, gotOwner, gotName, err := splitOrgRepo(tt.orgRepo)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantNoun, gotNoun)
			assert.Equal(t, tt.wantOwner, gotOwner)
			assert.Equal(t, tt.wantName, gotName)
		})
	}
}

func TestBitbucketServer_CreatePullRequest(t *testing.T) {
	tests := map[string]struct {
		opts     *PullRequestOptions
		wantURL  string
		wantErr  string
		beforeFn func(t *testing.T, c *mocks.MockHttpClient)
	}{
		"success": {
			opts: &PullRequestOptions{
				Owner:       "PROJECT",
				Repo:        "repo",
				Title:       "test PR",
				Description: "test description",
				Head:        "feature-branch",
				Base:        "main",
			},
			wantURL: "http://bitbucket.example.com/projects/PROJECT/repos/repo/pull-requests/1",
			beforeFn: func(t *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).DoAndReturn(
					func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "/rest/api/1.0/projects/PROJECT/repos/repo/pull-requests", req.URL.Path)

						// 验证请求体
						var reqBody map[string]interface{}
						body, _ := io.ReadAll(req.Body)
						json.Unmarshal(body, &reqBody)

						assert.Equal(t, "test PR", reqBody["title"])
						assert.Equal(t, "test description", reqBody["description"])
						assert.Equal(t, "refs/heads/feature-branch", reqBody["fromRef"].(map[string]interface{})["id"])
						assert.Equal(t, "refs/heads/main", reqBody["toRef"].(map[string]interface{})["id"])

						// 返回成功响应
						resp := &bitbucketServerPullRequestResponse{
							ID:    1,
							Title: "test PR",
							Links: struct {
								Self []struct {
									Href string `json:"href"`
								} `json:"self"`
							}{
								Self: []struct {
									Href string `json:"href"`
								}{
									{Href: "http://bitbucket.example.com/projects/PROJECT/repos/repo/pull-requests/1"},
								},
							},
						}
						return &http.Response{
							StatusCode: http.StatusCreated,
							Body:       createBody(resp),
						}, nil
					})
			},
		},
		"error": {
			opts: &PullRequestOptions{
				Owner: "invalid",
			},
			wantErr: "failed to create pull request",
			beforeFn: func(t *testing.T, c *mocks.MockHttpClient) {
				c.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Times(1).Return(
					&http.Response{
						StatusCode: http.StatusBadRequest,
						Body: createBody(errorBody{
							Errors: []bbError{
								{Message: "failed to create pull request"},
							},
						}),
					}, nil)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockClient := mocks.NewMockHttpClient(ctrl)

			if tt.beforeFn != nil {
				tt.beforeFn(t, mockClient)
			}

			bbs := &bitbucketServer{
				baseURL: baseURL(),
				c:       mockClient,
				opts:    providerOptions,
			}

			gotURL, err := bbs.CreatePullRequest(context.Background(), tt.opts)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantURL, gotURL)
		})
	}
}
