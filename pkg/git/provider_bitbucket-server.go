package git

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"

	"github.com/squidflow/service/pkg/util"
)

//go:generate mockgen -destination=./bitbucket-server/mocks/httpClient.go -package=mocks -source=./provider_bitbucket-server.go HttpClient

type (
	HttpClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	bitbucketServer struct {
		baseURL *url.URL
		c       HttpClient
		opts    *ProviderOptions
	}

	bbError struct {
		Context       string `json:"context"`
		Message       string `json:"message"`
		ExceptionName string `json:"exceptionName"`
	}

	errorBody struct {
		Errors []bbError `json:"errors"`
	}

	createRepoBody struct {
		Name          string `json:"name"`
		Scm           string `json:"scm"`
		DefaultBranch string `json:"defaultBranch"`
		Public        bool   `json:"public"`
	}

	Link struct {
		Name string `json:"name"`
		Href string `json:"href"`
	}

	Links struct {
		Clone []Link `json:"clone"`
	}

	repoResponse struct {
		Slug          string `json:"slug"`
		Name          string `json:"name"`
		Id            int32  `json:"id"`
		DefaultBranch string `json:"defaultBranch"`
		Public        bool   `json:"public"`
		Links         Links  `json:"links"`
	}

	userResponse struct {
		Slug         string `json:"slug"`
		Name         string `json:"name"`
		DisplayName  string `json:"displayName"`
		EmailAddress string `json:"emailAddress"`
	}
	refBody struct {
		Id string `json:"id"`
	}

	pullRequestResponse struct {
		ID    int   `json:"id"`
		Links Links `json:"links"`
	}
)

// Some tranditional company use self hosted bitbucket-server (atlassian tech stack)
const BitbucketServer = "bitbucket-server"

var (
	orgRepoReg = regexp.MustCompile("^scm/(~)?([^/]*)/([^/]*)$")
)

func newBitbucketServer(opts *ProviderOptions) (Provider, error) {
	host, _, _, _, _, _, _ := util.ParseGitUrl(opts.RepoURL)
	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	httpClient.Transport, err = DefaultTransportWithCa(opts.Auth.CertFile)
	if err != nil {
		return nil, err
	}

	g := &bitbucketServer{
		baseURL: baseURL,
		c:       httpClient,
		opts:    opts,
	}

	return g, nil
}

func (bbs *bitbucketServer) CreateRepository(ctx context.Context, orgRepo string) (defaultBranch string, err error) {
	noun, owner, name, err := splitOrgRepo(orgRepo)
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("%s/%s/repos", noun, owner)
	repo := &repoResponse{}
	err = bbs.requestRest(ctx, http.MethodPost, path, &createRepoBody{
		Name: name,
		Scm:  "git",
	}, repo)
	if err != nil {
		return "", err
	}

	return repo.DefaultBranch, nil
}

func (bbs *bitbucketServer) GetDefaultBranch(ctx context.Context, orgRepo string) (string, error) {
	noun, owner, name, err := splitOrgRepo(orgRepo)
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("%s/%s/repos/%s", noun, owner, name)
	repo := &repoResponse{}
	err = bbs.requestRest(ctx, http.MethodGet, path, nil, repo)
	if err != nil {
		return "", err
	}

	defaultBranch := repo.DefaultBranch
	if defaultBranch == "" {
		// fallback in case server response does not include the value at all
		// in both 6.10 and 8.2 i never actually got it in the response, and HAD to use this fallback
		defaultBranch = "master"
	}

	return defaultBranch, nil
}

func (bbs *bitbucketServer) GetAuthor(ctx context.Context) (username, email string, err error) {
	userSlug, err := bbs.whoAmI(ctx)
	if err != nil {
		err = fmt.Errorf("failed getting current user's slug: %w", err)
		return
	}

	user, err := bbs.getUser(ctx, userSlug)
	if err != nil {
		err = fmt.Errorf("failed getting current user: %w", err)
		return
	}

	username = user.DisplayName
	if username == "" {
		username = user.Name
	}

	email = user.EmailAddress
	if email == "" {
		email = user.Slug
	}

	return
}

func (bbs *bitbucketServer) whoAmI(ctx context.Context) (string, error) {
	data, err := bbs.request(ctx, http.MethodGet, "/plugins/servlet/applinks/whoami", nil)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (bbs *bitbucketServer) getUser(ctx context.Context, userSlug string) (*userResponse, error) {
	path := "users/" + userSlug
	user := &userResponse{}
	err := bbs.requestRest(ctx, http.MethodGet, path, nil, user)
	return user, err
}

func (bbs *bitbucketServer) requestRest(ctx context.Context, method, urlPath string, body interface{}, res interface{}) error {
	restPath := path.Join("rest/api/1.0", urlPath)
	data, err := bbs.request(ctx, method, restPath, body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, res)
}

func (bbs *bitbucketServer) request(ctx context.Context, method, urlPath string, body interface{}) ([]byte, error) {
	var err error

	urlClone := *bbs.baseURL
	urlClone.Path = path.Join(urlClone.Path, urlPath)
	bodyStr := []byte{}
	if body != nil {
		bodyStr, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, urlClone.String(), bytes.NewBuffer(bodyStr))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+bbs.opts.Auth.Password)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := bbs.c.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read from response body: %w", err)
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		error := &errorBody{}
		err = json.Unmarshal(data, error)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshalling error body \"%s\". error: %w", data, err)
		}

		return nil, errors.New(error.Errors[0].Message)
	}

	return data, nil
}

type bitbucketServerPullRequestResponse struct {
	ID          int    `json:"id"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Open        bool   `json:"open"`
	Closed      bool   `json:"closed"`
	CreatedDate int    `json:"createdDate"`
	UpdatedDate int    `json:"updatedDate"`
	FromRef     struct {
		ID           string `json:"id"`
		DisplayID    string `json:"displayId"`
		LatestCommit string `json:"latestCommit"`
		Type         string `json:"type"`
		Repository   struct {
			Slug          string `json:"slug"`
			ID            int    `json:"id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			HierarchyID   string `json:"hierarchyId"`
			ScmID         string `json:"scmId"`
			State         string `json:"state"`
			StatusMessage string `json:"statusMessage"`
			Forkable      bool   `json:"forkable"`
			Project       struct {
				Key         string `json:"key"`
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Public      bool   `json:"public"`
				Type        string `json:"type"`
				Links       struct {
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				} `json:"links"`
			} `json:"project"`
			Public bool `json:"public"`
			Links  struct {
				Clone []struct {
					Href string `json:"href"`
					Name string `json:"name"`
				} `json:"clone"`
				Self []struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"repository"`
	} `json:"fromRef"`
	ToRef struct {
		ID           string `json:"id"`
		DisplayID    string `json:"displayId"`
		LatestCommit string `json:"latestCommit"`
		Type         string `json:"type"`
		Repository   struct {
			Slug          string `json:"slug"`
			ID            int    `json:"id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			HierarchyID   string `json:"hierarchyId"`
			ScmID         string `json:"scmId"`
			State         string `json:"state"`
			StatusMessage string `json:"statusMessage"`
			Forkable      bool   `json:"forkable"`
			Project       struct {
				Key         string `json:"key"`
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Public      bool   `json:"public"`
				Type        string `json:"type"`
				Links       struct {
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				} `json:"links"`
			} `json:"project"`
			Public bool `json:"public"`
			Links  struct {
				Clone []struct {
					Href string `json:"href"`
					Name string `json:"name"`
				} `json:"clone"`
				Self []struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"repository"`
	} `json:"toRef"`
	Locked bool `json:"locked"`
	Author struct {
		User struct {
			Name         string `json:"name"`
			EmailAddress string `json:"emailAddress"`
			ID           int    `json:"id"`
			DisplayName  string `json:"displayName"`
			Active       bool   `json:"active"`
			Slug         string `json:"slug"`
			Type         string `json:"type"`
		} `json:"user"`
		Role     string `json:"role"`
		Approved bool   `json:"approved"`
		Status   string `json:"status"`
	} `json:"author"`
	Reviewers []struct {
		User struct {
			Name         string `json:"name"`
			EmailAddress string `json:"emailAddress"`
			ID           int    `json:"id"`
			DisplayName  string `json:"displayName"`
			Active       bool   `json:"active"`
			Slug         string `json:"slug"`
			Type         string `json:"type"`
		} `json:"user"`
		LastReviewedCommit string `json:"lastReviewedCommit"`
		Role               string `json:"role"`
		Approved           bool   `json:"approved"`
		Status             string `json:"status"`
	} `json:"reviewers"`
	Participants []any `json:"participants"`
	Links        struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

// for more details, see:
// https://docs.atlassian.com/bitbucket-server/rest/7.21.0/bitbucket-rest.html#idp301
func (bbs *bitbucketServer) CreatePullRequest(ctx context.Context, opts *PullRequestOptions) (string, error) {
	// construct the request body
	prRequest := struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		Open        bool   `json:"open"`
		Closed      bool   `json:"closed"`
		FromRef     struct {
			ID         string `json:"id"`
			Repository struct {
				Slug    string `json:"slug"`
				Project struct {
					Key string `json:"key"`
				} `json:"project"`
			} `json:"repository"`
		} `json:"fromRef"`
		ToRef struct {
			ID         string `json:"id"`
			Repository struct {
				Slug    string `json:"slug"`
				Project struct {
					Key string `json:"key"`
				} `json:"project"`
			} `json:"repository"`
		} `json:"toRef"`
		Locked bool `json:"locked"`
	}{
		Title:       opts.Title,
		Description: opts.Description,
		State:       "OPEN",
		Open:        true,
		Closed:      false,
		Locked:      false,
	}

	// set the source branch information
	prRequest.FromRef.ID = fmt.Sprintf("refs/heads/%s", opts.Head)
	prRequest.FromRef.Repository.Slug = opts.Repo
	prRequest.FromRef.Repository.Project.Key = opts.Owner

	// set the target branch information
	prRequest.ToRef.ID = fmt.Sprintf("refs/heads/%s", opts.Base)
	prRequest.ToRef.Repository.Slug = opts.Repo
	prRequest.ToRef.Repository.Project.Key = opts.Owner

	// send the request
	path := fmt.Sprintf("projects/%s/repos/%s/pull-requests", opts.Owner, opts.Repo)
	var prResponse bitbucketServerPullRequestResponse

	if err := bbs.requestRest(ctx, http.MethodPost, path, prRequest, &prResponse); err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	// return the PR URL
	if len(prResponse.Links.Self) > 0 {
		return prResponse.Links.Self[0].Href, nil
	}

	return "", fmt.Errorf("no pull request URL found in response")
}

func splitOrgRepo(orgRepo string) (noun, owner, name string, err error) {
	split := orgRepoReg.FindStringSubmatch(orgRepo)
	if len(split) == 0 {
		err = fmt.Errorf("invalid Bitbucket url \"%s\" - must be in the form of \"scm/[~]project-or-username/repo-name\"", orgRepo)
		return
	}

	noun = "projects"
	if split[1] == "~" {
		noun = "users"
	}

	owner = split[2]
	name = split[3]
	return
}
