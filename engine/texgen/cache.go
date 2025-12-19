// Package texgen provides dimension-aware texture generation for interiors.
package texgen

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Cache provides multi-level caching for generated textures.
// Level 1: Memory LRU cache for fast access to recent textures
// Level 2: Disk cache with hash-based filenames for persistence
type Cache struct {
	mu sync.RWMutex

	// Memory cache (LRU)
	memCache    map[string]*list.Element
	memOrder    *list.List
	memCapacity int

	// Disk cache
	diskDir string
}

// cacheEntry represents an entry in the memory cache.
type cacheEntry struct {
	key  string
	path string
}

// NewCache creates a new texture cache.
// memCapacity is the maximum number of entries in the memory cache.
// diskDir is the directory for the disk cache.
func NewCache(memCapacity int, diskDir string) (*Cache, error) {
	// Ensure disk cache directory exists
	if err := os.MkdirAll(diskDir, 0755); err != nil {
		return nil, err
	}

	return &Cache{
		memCache:    make(map[string]*list.Element),
		memOrder:    list.New(),
		memCapacity: memCapacity,
		diskDir:     diskDir,
	}, nil
}

// Get retrieves a texture path from the cache.
// Returns (path, true) if found, ("", false) if not.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()

	// Check memory cache first
	if elem, ok := c.memCache[key]; ok {
		c.mu.RUnlock()
		// Move to front (LRU)
		c.mu.Lock()
		c.memOrder.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		c.mu.Unlock()
		log.Printf("[cache] Memory hit: %s", key)
		return entry.path, true
	}
	c.mu.RUnlock()

	// Check disk cache
	diskPath := c.diskPath(key)
	if _, err := os.Stat(diskPath); err == nil {
		// Found on disk, add to memory cache
		c.mu.Lock()
		c.addToMemory(key, diskPath)
		c.mu.Unlock()
		log.Printf("[cache] Disk hit: %s", key)
		return diskPath, true
	}

	return "", false
}

// Put stores a texture in the cache.
// The sourcePath should be the path to an existing texture file.
// If sourcePath is in the cache directory, it's used directly.
// Otherwise, it's copied to the cache directory.
func (c *Cache) Put(key, sourcePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Determine the destination path
	destPath := c.diskPath(key)

	// If source is not already in cache dir, copy it
	if sourcePath != destPath {
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}
		log.Printf("[cache] Copied to disk cache: %s -> %s", sourcePath, destPath)
	}

	// Add to memory cache
	c.addToMemory(key, destPath)

	return nil
}

// addToMemory adds an entry to the memory cache (must hold lock).
func (c *Cache) addToMemory(key, path string) {
	// Check if already in memory
	if elem, ok := c.memCache[key]; ok {
		c.memOrder.MoveToFront(elem)
		return
	}

	// Evict oldest if at capacity
	for c.memOrder.Len() >= c.memCapacity {
		oldest := c.memOrder.Back()
		if oldest != nil {
			entry := oldest.Value.(*cacheEntry)
			delete(c.memCache, entry.key)
			c.memOrder.Remove(oldest)
			log.Printf("[cache] Evicted from memory: %s", entry.key)
		}
	}

	// Add new entry
	entry := &cacheEntry{key: key, path: path}
	elem := c.memOrder.PushFront(entry)
	c.memCache[key] = elem
}

// diskPath returns the disk cache path for a key.
// Uses SHA256 hash to create filesystem-safe filenames.
func (c *Cache) diskPath(key string) string {
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	// Use first 16 chars of hash + .png extension
	return filepath.Join(c.diskDir, hashStr[:16]+".png")
}

// Clear removes all entries from both memory and disk cache.
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear memory cache
	c.memCache = make(map[string]*list.Element)
	c.memOrder = list.New()

	// Clear disk cache
	entries, err := os.ReadDir(c.diskDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(c.diskDir, entry.Name())
			if err := os.Remove(path); err != nil {
				log.Printf("[cache] Warning: failed to remove %s: %v", path, err)
			}
		}
	}

	log.Printf("[cache] Cleared all caches")
	return nil
}

// Stats returns cache statistics.
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Count disk entries
	diskCount := 0
	entries, err := os.ReadDir(c.diskDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				diskCount++
			}
		}
	}

	return CacheStats{
		MemoryEntries:  c.memOrder.Len(),
		MemoryCapacity: c.memCapacity,
		DiskEntries:    diskCount,
		DiskDir:        c.diskDir,
	}
}

// CacheStats holds cache statistics.
type CacheStats struct {
	MemoryEntries  int
	MemoryCapacity int
	DiskEntries    int
	DiskDir        string
}
