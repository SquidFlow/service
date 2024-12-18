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
		mockRepo.EXPECT().Worktree().Return(nil, nil).AnyTimes()

		gitRepo := &mockGitRepo{
			MockRepository: mockRepo,
		}

		// test set
		fs := memfs.New()
		testURL := "https://github.com/test/repo.git"
		cache.Set(testURL, gitRepo, fs)

		// test get
		repo, filesystem, exists := cache.Get(testURL, false)
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

		// Create mock objects
		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().Worktree().Return(nil, nil).AnyTimes()

		gitRepo := &mockGitRepo{
			MockRepository: mockRepo,
		}

		fs := memfs.New()

		// Add first repo
		url1 := "https://github.com/test/repo1.git"
		cache.Set(url1, gitRepo, fs)
		assert.Equal(t, 1, len(cache.cache))

		// Add second repo, should evict first one due to capacity limit
		url2 := "https://github.com/test/repo2.git"
		cache.Set(url2, gitRepo, fs)
		assert.Equal(t, 1, len(cache.cache))

		// Verify first repo was evicted
		_, _, exists := cache.Get(url1, false)
		assert.False(t, exists, "First repo should have been evicted")

		// Verify second repo exists
		repo, filesystem, exists := cache.Get(url2, false)
		assert.True(t, exists, "Second repo should exist")
		assert.NotNil(t, repo)
		assert.NotNil(t, filesystem)

		// Test cache entry expiration
		time.Sleep(time.Millisecond * 100)
		cache.maxAge = time.Millisecond * 50
		_, _, exists = cache.Get(url2, false)
		assert.False(t, exists, "Entry should have expired")
	})
}
