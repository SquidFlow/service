package git

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-git/go-billy/v5"
	gg "github.com/go-git/go-git/v5"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/log"
)

type RepositoryCache interface {
	Get(url string) (Repository, fs.FS, bool)
	Set(url string, repo Repository, filesystem billy.Filesystem)
	SyncRepo(repository Repository) error
}

type repositoryCache struct {
	mu       sync.RWMutex
	cache    map[string]*repositoryCacheEntry
	maxAge   time.Duration
	capacity int
}

type repositoryCacheEntry struct {
	repo     Repository
	fs       billy.Filesystem
	lastUsed time.Time
	lastSync time.Time
}

var (
	defaultRepositoryCache *repositoryCache
	repositoryCacheOnce    sync.Once
)

func getRepositoryCache() *repositoryCache {
	repositoryCacheOnce.Do(func() {
		defaultRepositoryCache = &repositoryCache{
			cache:    make(map[string]*repositoryCacheEntry),
			maxAge:   30 * time.Minute,
			capacity: 10,
		}
		go defaultRepositoryCache.cleanup()
	})
	return defaultRepositoryCache
}

func (c *repositoryCache) get(url string) (Repository, fs.FS, bool) {
	c.mu.RLock()
	entry, exists := c.cache[url]
	c.mu.RUnlock()

	if !exists {
		return nil, nil, false
	}

	if time.Since(entry.lastUsed) > c.maxAge {
		c.mu.Lock()
		delete(c.cache, url)
		c.mu.Unlock()
		return nil, nil, false
	}

	needSync := time.Since(entry.lastSync) > 5*time.Minute

	if needSync {
		log.G().WithFields(log.Fields{
			"url": url,
		}).Debug("Syncing cached repository")

		if err := c.SyncRepo(entry.repo); err != nil {
			log.G().WithError(err).Error("Failed to sync repository")
			c.mu.Lock()
			delete(c.cache, url)
			c.mu.Unlock()
			return nil, nil, false
		}

		c.mu.Lock()
		entry.lastSync = time.Now()
		c.mu.Unlock()
	}

	c.mu.Lock()
	entry.lastUsed = time.Now()
	c.mu.Unlock()

	return entry.repo, fs.Create(entry.fs), true
}

func (c *repositoryCache) set(url string, repo Repository, filesystem billy.Filesystem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) >= c.capacity {
		var oldestKey string
		var oldestTime time.Time
		first := true

		for k, v := range c.cache {
			if first || v.lastUsed.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.lastUsed
				first = false
			}
		}
		delete(c.cache, oldestKey)
	}

	c.cache[url] = &repositoryCacheEntry{
		repo:     repo,
		fs:       filesystem,
		lastUsed: time.Now(),
		lastSync: time.Now(),
	}
}

func (c *repositoryCache) cleanup() {
	ticker := time.NewTicker(c.maxAge / 2)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for url, entry := range c.cache {
			if now.Sub(entry.lastUsed) > c.maxAge {
				log.G().WithFields(log.Fields{
					"url": url,
					"age": now.Sub(entry.lastUsed),
				}).Debug("Removing expired cache entry")
				delete(c.cache, url)
			}
		}
		c.mu.Unlock()
	}
}

type GitRepository interface {
	Repository
	Worktree() (gg.Worktree, error)
}

func (c *repositoryCache) SyncRepo(repository Repository) error {
	r, ok := repository.(GitRepository)
	if !ok {
		return fmt.Errorf("repository does not implement required methods")
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Pull(&gg.PullOptions{
		RemoteName: "origin",
		Force:      true,
	})

	if err == gg.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}
