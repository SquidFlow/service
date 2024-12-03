package git

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-git/go-billy/v5"
	gg "github.com/go-git/go-git/v5"
	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/git/gogit"
	"github.com/squidflow/service/pkg/log"
)

type repositoryCache struct {
	mu       sync.RWMutex
	cache    map[string]*repositoryCacheEntry
	maxAge   time.Duration
	capacity int
}

type repositoryCacheEntry struct {
	repo     gogit.Repository
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
		log.G().Debug("Initializing repository cache")
		defaultRepositoryCache = &repositoryCache{
			cache:    make(map[string]*repositoryCacheEntry),
			maxAge:   30 * time.Minute,
			capacity: 10,
		}
		go defaultRepositoryCache.cleanup()
	})
	return defaultRepositoryCache
}

func (c *repositoryCache) get(url string, pull bool) (gogit.Repository, fs.FS, bool) {
	log.G().WithFields(log.Fields{
		"url":        url,
		"cache_size": len(c.cache),
		"pull":       pull,
	}).Debug("Trying to get from cache")

	c.mu.RLock()
	entry, exists := c.cache[url]
	c.mu.RUnlock()

	if !exists {
		log.G().WithField("url", url).Debug("Cache miss - entry not found")
		return nil, nil, false
	}

	if time.Since(entry.lastUsed) > c.maxAge {
		c.mu.Lock()
		delete(c.cache, url)
		c.mu.Unlock()
		log.G().WithFields(log.Fields{
			"url": url,
			"age": time.Since(entry.lastUsed),
		}).Debug("Cache miss - entry expired")
		return nil, nil, false
	}

	needSync := pull || time.Since(entry.lastSync) > 5*time.Minute
	if needSync {
		log.G().WithFields(log.Fields{
			"url":        url,
			"last_sync":  time.Since(entry.lastSync),
			"force_sync": true,
		}).Debug("Syncing cached repository")

		w, err := entry.repo.Worktree()
		if err != nil {
			log.G().WithError(err).Error("Failed to get worktree")
			return nil, nil, false
		}

		err = w.Pull(&gg.PullOptions{
			RemoteName: "origin",
			Force:      true,
		})

		if err != nil && err != gg.NoErrAlreadyUpToDate {
			log.G().WithError(err).Error("Failed to sync repository")
			return nil, nil, false
		}

		c.mu.Lock()
		entry.lastSync = time.Now()
		c.mu.Unlock()
	}

	c.mu.Lock()
	entry.lastUsed = time.Now()
	c.mu.Unlock()

	filesystem := fs.Create(entry.fs)

	log.G().WithFields(log.Fields{
		"url":       url,
		"cache_hit": true,
	}).Debug("Cache hit")

	return entry.repo, filesystem, true
}

func (c *repositoryCache) set(url string, repo gogit.Repository, filesystem billy.Filesystem) {
	log.G().WithFields(log.Fields{
		"url":        url,
		"repo_type":  fmt.Sprintf("%T", repo),
		"cache_size": len(c.cache),
	}).Debug("Setting cache entry")

	c.mu.Lock()
	defer c.mu.Unlock()

	// check if the entry already exists
	if existing, exists := c.cache[url]; exists {
		log.G().WithField("url", url).Debug("Updating existing cache entry")
		existing.repo = repo
		existing.fs = filesystem
		existing.lastUsed = time.Now()
		existing.lastSync = time.Now()
		return
	}

	// if the cache is full, evict the oldest entry
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
		log.G().WithFields(log.Fields{
			"evicted_key": oldestKey,
			"age":         time.Since(oldestTime),
		}).Debug("Evicting cache entry")
		delete(c.cache, oldestKey)
	}

	c.cache[url] = &repositoryCacheEntry{
		repo:     repo,
		fs:       filesystem,
		lastUsed: time.Now(),
		lastSync: time.Now(),
	}

	log.G().WithFields(log.Fields{
		"url":        url,
		"cache_size": len(c.cache),
	}).Debug("New cache entry set")
}

func (c *repositoryCache) cleanup() {
	ticker := time.NewTicker(c.maxAge / 2)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		beforeCount := len(c.cache)
		for url, entry := range c.cache {
			if now.Sub(entry.lastUsed) > c.maxAge {
				log.G().WithFields(log.Fields{
					"url": url,
					"age": now.Sub(entry.lastUsed),
				}).Debug("Removing expired cache entry")
				delete(c.cache, url)
			}
		}
		afterCount := len(c.cache)
		if beforeCount != afterCount {
			log.G().WithFields(log.Fields{
				"before": beforeCount,
				"after":  afterCount,
			}).Debug("Cache cleanup completed")
		}
		c.mu.Unlock()
	}
}
