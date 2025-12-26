package analytics

import (
	"sync"
	"time"
)

// cacheEntry holds cached stats and metadata
type cacheEntry struct {
	stats       []Stats
	lastRefresh time.Time
}

// statsCache provides thread-safe caching for analytics statistics
type statsCache struct {
	mu        sync.RWMutex
	perFile   map[string]*cacheEntry // key: profileName
	perPath   map[string]*cacheEntry // key: profileName
	ttl       time.Duration          // cache time-to-live
}

// newStatsCache creates a new statistics cache with the specified TTL
func newStatsCache(ttl time.Duration) *statsCache {
	return &statsCache{
		perFile: make(map[string]*cacheEntry),
		perPath: make(map[string]*cacheEntry),
		ttl:     ttl,
	}
}

// getPerFile retrieves cached per-file stats if available and fresh
func (c *statsCache) getPerFile(profileName string) ([]Stats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.perFile[profileName]
	if !exists {
		return nil, false
	}

	// Check if cache is still fresh
	if time.Since(entry.lastRefresh) > c.ttl {
		return nil, false
	}

	return entry.stats, true
}

// setPerFile stores per-file stats in cache
func (c *statsCache) setPerFile(profileName string, stats []Stats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.perFile[profileName] = &cacheEntry{
		stats:       stats,
		lastRefresh: time.Now(),
	}
}

// getPerPath retrieves cached per-path stats if available and fresh
func (c *statsCache) getPerPath(profileName string) ([]Stats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.perPath[profileName]
	if !exists {
		return nil, false
	}

	// Check if cache is still fresh
	if time.Since(entry.lastRefresh) > c.ttl {
		return nil, false
	}

	return entry.stats, true
}

// setPerPath stores per-path stats in cache
func (c *statsCache) setPerPath(profileName string, stats []Stats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.perPath[profileName] = &cacheEntry{
		stats:       stats,
		lastRefresh: time.Now(),
	}
}

// invalidate clears all cached data
func (c *statsCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.perFile = make(map[string]*cacheEntry)
	c.perPath = make(map[string]*cacheEntry)
}

// invalidateProfile clears cached data for a specific profile
func (c *statsCache) invalidateProfile(profileName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.perFile, profileName)
	delete(c.perPath, profileName)
}
