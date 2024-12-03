package git

import (
	"context"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/golang/mock/gomock"
	"github.com/squidflow/service/pkg/git/gogit/mocks"
	"github.com/stretchr/testify/assert"
)

type mockGitRepo struct {
	*mocks.MockRepository
	worktree *mocks.MockWorktree
}

func (m *mockGitRepo) Worktree() (mocks.MockWorktree, error) {
	return *m.worktree, nil
}

func (m *mockGitRepo) CurrentBranch() (string, error) {
	return "main", nil
}

func (m *mockGitRepo) Persist(ctx context.Context, opts *PushOptions) (string, error) {
	return "commit-id", nil
}

func TestRepositoryCache_Interface(t *testing.T) {
	t.Run("test get and set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cache := &repositoryCache{
			cache:    make(map[string]*repositoryCacheEntry),
			maxAge:   time.Minute,
			capacity: 2,
		}

		// create mock objects
		mockRepo := mocks.NewMockRepository(ctrl)
		mockWt := mocks.NewMockWorktree(ctrl)
		gitRepo := &mockGitRepo{
			MockRepository: mockRepo,
			worktree:       mockWt,
		}

		// test set
		fs := memfs.New()
		testURL := "https://github.com/test/repo.git"
		cache.set(testURL, gitRepo, fs)

		// test get
		repo, filesystem, exists := cache.get(testURL)
		assert.True(t, exists)
		assert.NotNil(t, repo)
		assert.NotNil(t, filesystem)
	})

	t.Run("test capacity limit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cache := &repositoryCache{
			cache:    make(map[string]*repositoryCacheEntry),
			maxAge:   time.Minute,
			capacity: 1,
		}

		mockRepo := mocks.NewMockRepository(ctrl)
		mockWt := mocks.NewMockWorktree(ctrl)
		gitRepo := &mockGitRepo{
			MockRepository: mockRepo,
			worktree:       mockWt,
		}

		fs := memfs.New()

		// add first repo
		url1 := "https://github.com/test/repo1.git"
		cache.set(url1, gitRepo, fs)
		assert.Equal(t, 1, len(cache.cache))

		// add second repo
		url2 := "https://github.com/test/repo2.git"
		cache.set(url2, gitRepo, fs)
		assert.Equal(t, 1, len(cache.cache))

		// verify first repo is removed
		_, _, exists := cache.get(url1)
		assert.False(t, exists)

		// verify second repo exists
		_, _, exists = cache.get(url2)
		assert.True(t, exists)
	})
}
