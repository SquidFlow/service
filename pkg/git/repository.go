package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git/gogit"
	"github.com/squidflow/service/pkg/log"
	"github.com/squidflow/service/pkg/store"
	"github.com/squidflow/service/pkg/util"

	billy "github.com/go-git/go-billy/v5"
	gg "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:generate mockgen -destination=./mocks/repository.go -package=mocks -source=./repository.go Repository
//go:generate mockgen -destination=./gogit/mocks/repository.go -package=mocks -source=./gogit/repo.go Repository
//go:generate mockgen -destination=./gogit/mocks/worktree.go -package=mocks -source=./gogit/worktree.go Worktree

type (
	// Repository represents a git repository
	Repository interface {
		// Persist runs add, commit and push to the repository default remote
		Persist(ctx context.Context, opts *PushOptions) (string, error)
		// CurrentBranch returns the name of the current branch
		CurrentBranch() (string, error)
	}

	AddFlagsOptions struct {
		FS               billy.Filesystem
		Prefix           string
		CreateIfNotExist bool
		// CloneForWrite if true will not allow 'ref' query param which is not
		// a branch name
		CloneForWrite bool
		Optional      bool
	}

	CloneOptions struct {
		Provider         string
		Repo             string
		Auth             Auth
		FS               fs.FS
		Progress         io.Writer
		CreateIfNotExist bool
		CloneForWrite    bool
		UpsertBranch     bool

		url        string
		revision   string
		submodules bool
		path       string
	}

	PushOptions struct {
		Provider       string
		AddGlobPattern string
		CommitMsg      string
		Progress       io.Writer
	}

	repo struct {
		gogit.Repository
		auth         Auth
		progress     io.Writer
		providerType string
		repoURL      string
	}
)

// Defaults
const (
	pushRetries        = 3
	failureBackoffTime = 3 * time.Second
)

// Errors
var (
	ErrNilOpts      = errors.New("options cannot be nil")
	ErrNoParse      = errors.New("must call Parse before using CloneOptions")
	ErrRepoNotFound = errors.New("git repository not found")
	ErrNoRemotes    = errors.New("no remotes in repository")

	// go-git functions (we mock those in tests)
	checkoutRef = func(r *repo, ref string) error {
		return r.checkoutRef(ref)
	}

	checkoutBranch = func(r *repo, branch string, upsertBranch bool) error {
		return r.checkoutBranch(branch, upsertBranch)
	}

	getProvider = func(providerType, repoURL string, auth *Auth) (Provider, error) {
		if providerType == "" {
			u, err := url.Parse(repoURL)
			if err != nil {
				return nil, err
			}

			providerType = strings.TrimSuffix(u.Hostname(), ".com")

			log.G().Warnf("--provider not specified, assuming provider from url: %s", providerType)
		}

		return newProvider(&ProviderOptions{
			Type:    providerType,
			Auth:    auth,
			RepoURL: repoURL,
		})
	}

	ggClone = func(ctx context.Context, s storage.Storer, worktree billy.Filesystem, o *gg.CloneOptions) (gogit.Repository, error) {
		return gg.CloneContext(ctx, s, worktree, o)
	}

	ggInitRepo = func(s storage.Storer, worktree billy.Filesystem) (gogit.Repository, error) {
		return gg.Init(s, worktree)
	}

	worktree = func(r gogit.Repository) (gogit.Worktree, error) {
		return r.Worktree()
	}
)

func AddFlags(cmd *cobra.Command, opts *AddFlagsOptions) *CloneOptions {
	co := &CloneOptions{
		FS:               fs.Create(opts.FS),
		CreateIfNotExist: opts.CreateIfNotExist,
		CloneForWrite:    opts.CloneForWrite,
	}

	if opts.Prefix != "" && !strings.HasSuffix(opts.Prefix, "-") {
		opts.Prefix += "-"
	}

	envPrefix := strings.ReplaceAll(strings.ToUpper(opts.Prefix), "-", "_")
	cmd.PersistentFlags().StringVar(&co.Auth.Password, opts.Prefix+"git-token", "", fmt.Sprintf("Your git provider api token [%sGIT_TOKEN]", envPrefix))
	cmd.PersistentFlags().StringVar(&co.Auth.Username, opts.Prefix+"git-user", "", fmt.Sprintf("Your git provider user name [%sGIT_USER] (not required in GitHub)", envPrefix))
	cmd.PersistentFlags().StringVar(&co.Auth.CertFile, opts.Prefix+"git-server-crt", "", fmt.Sprint("Git Server certificate file", envPrefix))
	cmd.PersistentFlags().StringVar(&co.Repo, opts.Prefix+"repo", "", fmt.Sprintf("Repository URL [%sGIT_REPO]", envPrefix))

	util.Die(viper.BindEnv(opts.Prefix+"git-token", envPrefix+"GIT_TOKEN"))
	util.Die(viper.BindEnv(opts.Prefix+"git-user", envPrefix+"GIT_USER"))
	util.Die(viper.BindEnv(opts.Prefix+"repo", envPrefix+"GIT_REPO"))

	if opts.Prefix == "" {
		cmd.Flag("git-token").Shorthand = "t"
		cmd.Flag("git-user").Shorthand = "u"
	}

	cmd.PersistentFlags().StringVar(&co.Provider, opts.Prefix+"provider", "", fmt.Sprintf("The git provider, one of: %v", strings.Join(Providers(), "|")))
	if !opts.CreateIfNotExist {
		util.Die(cmd.PersistentFlags().MarkHidden(opts.Prefix + "provider"))
	}

	if opts.CloneForWrite {
		cmd.PersistentFlags().BoolVarP(&co.UpsertBranch, opts.Prefix+"upsert-branch", "b", false, "If true will try to checkout the specified branch and create it if it doesn't exist")
	}

	if !opts.Optional {
		util.Die(cmd.MarkPersistentFlagRequired(opts.Prefix + "git-token"))
		util.Die(cmd.MarkPersistentFlagRequired(opts.Prefix + "repo"))
	}

	cmd.PersistentFlags().BoolVar(&co.submodules, opts.Prefix+"submodules", false, "Clone with submodules")

	return co
}

func (o *CloneOptions) Parse() {
	var (
		host    string
		orgRepo string
		suffix  string
	)

	host, orgRepo, o.path, o.revision, o.submodules, suffix, _ = util.ParseGitUrl(o.Repo)
	o.url = host + orgRepo + suffix

	if o.Auth.Username == "" {
		o.Auth.Username = store.Default.GitHubUsername
	}
}

func (o *CloneOptions) Revision() string {
	return o.revision
}

func (o *CloneOptions) SetRevision(revision string) {
	o.revision = revision
}

func (o *CloneOptions) URL() string {
	return o.url
}

func (o *CloneOptions) Path() string {
	return o.path
}

func (o *CloneOptions) GetRepo(ctx context.Context) (Repository, fs.FS, error) {
	if o == nil {
		return nil, nil, ErrNilOpts
	}

	if o.url == "" {
		return nil, nil, ErrNoParse
	}

	// Add debug logging to check cache key
	log.G().WithFields(log.Fields{
		"url":          o.url,
		"repo":         o.Repo,
		"path":         o.path,
		"write_mode":   o.CloneForWrite,
	}).Debug("Trying to get repo from cache")

	// Try to get from cache first
	cachedRepo, filesystem, exists := getRepositoryCache().get(o.url, o.CloneForWrite)
	if exists {
		log.G().WithField("url", o.url).Debug("Cache hit")
		// Create a new repo instance with the cached repository
		wrappedRepo := &repo{
			Repository:   cachedRepo,
			auth:        o.Auth,
			progress:    o.Progress,
			repoURL:     o.Repo,
			providerType: o.Provider,
		}

		// For write operations, validate permissions
		if o.CloneForWrite {
			if err := validateRepoWritePermission(ctx, wrappedRepo); err != nil {
				return nil, nil, fmt.Errorf("failed to validate write permission: %w", err)
			}
		}

		return wrappedRepo, filesystem, nil
	}

	log.G().WithField("url", o.url).Debug("Cache miss")

	// Cache miss, perform clone
	newRepo, err := clone(ctx, o)
	if err != nil {
		switch err {
		case transport.ErrRepositoryNotFound:
			if !o.CreateIfNotExist {
				return nil, nil, err
			}

			log.G().Infof("repository '%s' was not found, trying to create it...", o.Repo)
			defaultBranch, err := createRepo(ctx, o)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create repository: %w", err)
			}

			newRepo, err = initRepo(ctx, o, defaultBranch)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to initialize repository: %w", err)
			}

		case transport.ErrEmptyRemoteRepository:
			log.G().Info("empty repository, initializing a new one with specified remote")
			newRepo, err = initRepo(ctx, o, "")
			if err != nil {
				return nil, nil, fmt.Errorf("failed to initialize repository: %w", err)
			}

		default:
			return nil, nil, err
		}
	}

	// Handle submodules if needed
	if o.submodules && newRepo != nil {
		if err := newRepo.cloneSubmodules(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed to clone submodules: %w", err)
		}
	}

	// Create filesystem
	bootstrapFS, err := o.FS.Chroot(o.path)
	if err != nil {
		return nil, nil, err
	}

	// Store in cache with original filesystem
	getRepositoryCache().set(o.url, newRepo, bootstrapFS)

	// Create fs.FS wrapper for return
	filesystem = fs.Create(bootstrapFS)

	return newRepo, filesystem, nil
}

var validateRepoWritePermission = func(ctx context.Context, r *repo) error {
	_, err := r.Persist(ctx, &PushOptions{
		CommitMsg: "chore: validating repository write permission",
	})

	if err != nil {
		return fmt.Errorf("failed pushing commit to repository: %w", err)
	}

	return nil
}

func (r *repo) Persist(ctx context.Context, opts *PushOptions) (string, error) {
	if opts == nil {
		return "", ErrNilOpts
	}

	progress := opts.Progress
	if progress == nil {
		progress = r.progress
	}

	cert, err := r.auth.GetCertificate()
	if err != nil {
		return "", fmt.Errorf("failed reading git certificate file: %w", err)
	}

	h, err := r.commit(ctx, opts)
	if err != nil {
		return "", err
	}

	for try := 0; try < pushRetries; try++ {
		err = r.PushContext(ctx, &gg.PushOptions{
			Auth:     getAuth(r.auth),
			Progress: progress,
			CABundle: cert,
		})
		if err == nil || !errors.Is(err, transport.ErrRepositoryNotFound) {
			break
		}

		log.G().WithFields(log.Fields{
			"retry": try,
			"err":   err.Error(),
		}).Warn("Failed to push to repository, trying again in 3 seconds...")

		time.Sleep(failureBackoffTime)
	}

	return h.String(), err
}

func (r *repo) CurrentBranch() (string, error) {
	ref, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("failed to resolve ref: %w", err)
	}

	return ref.Name().Short(), nil
}

func (r *repo) commit(ctx context.Context, opts *PushOptions) (*plumbing.Hash, error) {
	var h plumbing.Hash

	author, err := r.getAuthor(ctx)
	if err != nil {
		return nil, err
	}

	w, err := worktree(r)
	if err != nil {
		return nil, err
	}

	addPattern := "."
	if opts.AddGlobPattern != "" {
		addPattern = opts.AddGlobPattern
	}

	if err := w.AddGlob(addPattern); err != nil {
		// allowing the glob pattern to not match any files, in case of add-all ("."), like with initBranch for example
		if addPattern != "." || err != gg.ErrGlobNoMatches {
			return nil, err
		}
	}

	h, err = w.Commit(opts.CommitMsg, &gg.CommitOptions{
		All:               true,
		Author:            author,
		AllowEmptyCommits: true,
	})
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func (r *repo) getAuthor(ctx context.Context) (*object.Signature, error) {
	cfg, err := r.ConfigScoped(config.SystemScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get gitconfig: %w", err)
	}

	username := cfg.User.Name
	email := cfg.User.Email

	if username == "" || email == "" {
		provider, _ := getProvider(r.providerType, r.repoURL, &r.auth)
		if provider != nil {
			username, email, err = provider.GetAuthor(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get author information: %w", err)
			}
		}
	}

	if username == "" || email == "" {
		return nil, fmt.Errorf("missing required author information in git config, make sure your git config contains a 'user.name' and 'user.email'")
	}

	return &object.Signature{
		Name:  username,
		Email: email,
		When:  time.Now(),
	}, nil
}

func (r *repo) getConfigDefaultBranch() (string, error) {
	cfg, err := r.ConfigScoped(config.SystemScope)
	if err != nil {
		return "", fmt.Errorf("failed to get gitconfig: %w", err)
	}

	defaultBranch := cfg.Init.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	return defaultBranch, nil
}

var clone = func(ctx context.Context, opts *CloneOptions) (*repo, error) {
	var (
		err            error
		r              gogit.Repository
		curPushRetries = pushRetries
	)

	if opts == nil {
		return nil, ErrNilOpts
	}

	progress := opts.Progress
	if progress == nil {
		progress = os.Stderr
	}

	cert, err := opts.Auth.GetCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed reading git certificate file: %w", err)
	}

	cloneOpts := &gg.CloneOptions{
		URL:               opts.url,
		Auth:              getAuth(opts.Auth),
		Depth:             1,
		Progress:          progress,
		CABundle:          cert,
		RecurseSubmodules: gg.DefaultSubmoduleRecursionDepth,
	}

	log.G().WithField("url", opts.url).Debug("cloning git repo")

	if opts.CreateIfNotExist {
		curPushRetries = 1 // no retries
	}

	for try := 0; try < curPushRetries; try++ {
		r, err = ggClone(ctx, memory.NewStorage(), opts.FS, cloneOpts)
		if bitbucketServerNotFound(err) {
			err = transport.ErrRepositoryNotFound
		}

		if err == nil || !errors.Is(err, transport.ErrRepositoryNotFound) {
			break
		}

		log.G().WithFields(log.Fields{
			"retry": try,
			"err":   err.Error(),
		}).Debug("Failed to clone repository, trying again in 3 seconds...")

		time.Sleep(failureBackoffTime)
	}

	if err != nil {
		return nil, err
	}

	repo := &repo{
		Repository:   r,
		auth:         opts.Auth,
		progress:     progress,
		providerType: opts.Provider,
		repoURL:      opts.Repo,
	}

	if opts.revision != "" {
		if opts.CloneForWrite {
			log.G().WithFields(log.Fields{
				"branch": opts.revision,
				"upsert": opts.UpsertBranch,
			}).Debug("Trying to checkout branch")

			if err := checkoutBranch(repo, opts.revision, opts.UpsertBranch); err != nil {
				return nil, err
			}
		} else {
			log.G().WithField("ref", opts.revision).Debug("Trying to checkout ref")

			if err := checkoutRef(repo, opts.revision); err != nil {
				return nil, err
			}
		}
	}

	return repo, nil
}

var createRepo = func(ctx context.Context, opts *CloneOptions) (defaultBranch string, err error) {
	provider, _ := getProvider(opts.Provider, opts.Repo, &opts.Auth)
	if provider == nil {
		return "", errors.New("failed creating repository - no git provider supplied")
	}

	_, orgRepo, _, _, _, _, _ := util.ParseGitUrl(opts.Repo)

	// It depends on the provider, but org repo strucure should at least contain org and repo name
	slc := util.CleanSliceWhiteSpaces(strings.Split(orgRepo, "/"))
	if len(slc) < 2 {
		return "", errors.New("repo name can't be empty")
	}

	return provider.CreateRepository(ctx, orgRepo)
}

func getDefaultRepoOptions(orgRepo string) (*CreateRepoOptions, error) {
	s := strings.Split(orgRepo, "/")
	if len(s) < 2 {
		return nil, fmt.Errorf("failed parsing organization and repo from '%s'", orgRepo)
	}

	owner := strings.Join(s[:len(s)-1], "/")
	name := s[len(s)-1]
	return &CreateRepoOptions{
		Owner:   owner,
		Name:    name,
		Private: true,
	}, nil
}

var initRepo = func(ctx context.Context, opts *CloneOptions, defaultBranch string) (*repo, error) {
	ggr, err := ggInitRepo(memory.NewStorage(), opts.FS)
	if err != nil {
		return nil, err
	}

	progress := opts.Progress
	if progress == nil {
		progress = os.Stderr
	}

	r := &repo{
		Repository:   ggr,
		progress:     progress,
		providerType: opts.Provider,
		repoURL:      opts.Repo,
		auth:         opts.Auth,
	}
	if err = r.addRemote("origin", opts.url); err != nil {
		return nil, err
	}

	if defaultBranch == "" {
		defaultBranch, err = r.getDefaultBranch(ctx, opts.Repo)
		if err != nil {
			return nil, err
		}
	}

	if defaultBranch != plumbing.Master.Short() {
		if err = fixDefaultBranch(ggr, defaultBranch); err != nil {
			return nil, fmt.Errorf("failed to set default branch in new repository. Error: %w", err)
		}
	}

	branchName := opts.revision
	if branchName == "" {
		branchName = defaultBranch
	}

	return r, r.initBranch(ctx, branchName)
}

func (r *repo) getDefaultBranch(ctx context.Context, repo string) (string, error) {
	_, orgRepo, _, _, _, _, _ := util.ParseGitUrl(repo)
	provider, err := getProvider(r.providerType, r.repoURL, &r.auth)
	if err != nil {
		return "", err
	}

	defaultBranch, err := provider.GetDefaultBranch(ctx, orgRepo)
	if err != nil {
		return "", fmt.Errorf("failed to get default branch from provider. Error: %w", err)
	}

	if defaultBranch == "" {
		defaultBranch, err = r.getConfigDefaultBranch()
		if err != nil {
			return "", fmt.Errorf("failed to get default branch from global config. Error: %w", err)
		}
	}

	return defaultBranch, nil
}

func (r *repo) checkoutBranch(branch string, upsertBranch bool) error {
	wt, err := worktree(r)
	if err != nil {
		return err
	}

	err = wt.Checkout(&gg.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != plumbing.ErrReferenceNotFound {
		return err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		return ErrNoRemotes
	}

	err = wt.Checkout(&gg.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName(remotes[0].Config().Name, branch),
	})
	if err != nil {
		if err == plumbing.ErrReferenceNotFound && upsertBranch {
			// no remote branch but create is true
			// so we will create a new local branch
			return wt.Checkout(&gg.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName(branch),
				Create: true,
			})

		}

		return err
	}

	// if succeeded to checkout to a remote branch with this name,
	// checkout to a local branch from the remote branch
	return wt.Checkout(&gg.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	})
}

func (r *repo) checkoutRef(ref string) error {
	hash, err := r.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return err
		}

		log.G().WithField("ref", ref).Debug("failed resolving ref, trying to resolve from remote branch")
		remotes, err := r.Remotes()
		if err != nil {
			return err
		}

		if len(remotes) == 0 {
			return ErrNoRemotes
		}

		remoteref := fmt.Sprintf("%s/%s", remotes[0].Config().Name, ref)
		hash, err = r.ResolveRevision(plumbing.Revision(remoteref))
		if err != nil {
			return err
		}
	}

	wt, err := worktree(r)
	if err != nil {
		return err
	}

	log.G().WithFields(log.Fields{
		"ref":  ref,
		"hash": hash.String(),
	}).Debug("checking out commit")
	return wt.Checkout(&gg.CheckoutOptions{
		Hash: *hash,
	})
}

func (r *repo) addRemote(name, url string) error {
	_, err := r.CreateRemote(&config.RemoteConfig{Name: name, URLs: []string{url}})
	return err
}

func (r *repo) initBranch(ctx context.Context, branchName string) error {
	_, err := r.commit(ctx, &PushOptions{
		CommitMsg: "init: first commit of h4-platform",
	})
	if err != nil {
		return fmt.Errorf("failed to commit while trying to initialize the branch. Error: %w", err)
	}

	b := plumbing.NewBranchReferenceName(branchName)
	_, err = r.Reference(b, true)
	create := false
	if err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return fmt.Errorf("failed to check if branch exist. Error: %w", err)
		}

		// error is ReferenceNotFound - we need to create the branch on checkout
		create = true
	}

	log.G().WithField("branch", b).Debug("checking out branch")

	w, err := worktree(r)
	if err != nil {
		return err
	}

	return w.Checkout(&gg.CheckoutOptions{
		Branch: b,
		Create: create,
	})
}

func getAuth(auth Auth) transport.AuthMethod {
	if auth.Password == "" {
		return nil
	}

	return &http.BasicAuth{
		Username: auth.Username,
		Password: auth.Password,
	}
}

// a hack to handle case where bitbucket-server returns http 200 when repo not found
// and go-git fails to understand it
func bitbucketServerNotFound(err error) bool {
	e, ok := err.(*packp.ErrUnexpectedData)
	if !ok {
		return false
	}

	return string(e.Data) == "ERR Repository not found\nThe requested repository does not exist, or you do not have permission to\naccess it."
}

func fixDefaultBranch(r gogit.Repository, defaultBranch string) error {
	rInstance, ok := r.(*gg.Repository)
	if !ok {
		return errors.New("failed casting repo from go-git")
	}

	if err := rInstance.Storer.RemoveReference(plumbing.Master); err != nil {
		return err
	}

	defRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName(defaultBranch))
	return rInstance.Storer.SetReference(defRef)
}

func (r *repo) cloneSubmodules(ctx context.Context) error {
	w, err := worktree(r)
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	log.G().WithField("repo", r.repoURL).Debugf("Cloning submodules of repository")

	subs, err := w.Submodules()
	if err != nil {
		return fmt.Errorf("failed to get submodules: %w", err)
	}

	log.G().Infof("Found %d submodules", len(subs))

	for _, sub := range subs {
		log.G().Debugf("Updating submodule: %s", sub.Config().Name)
		_, err := sub.Repository()
		if err != nil {
			return fmt.Errorf("failed to get submodule repository: %w", err)
		}
		log.G().Infof("Submodule repository: %s", sub.Config().Name)
	}

	return nil
}
