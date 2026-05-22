package cache

import (
	"sync"

	"neko_backend/models"
)

// KeyCache is an in-process read-through cache for license keys.
// Write operations must call Delete to invalidate the entry.
type KeyCache struct {
	mu    sync.RWMutex
	items map[string]*models.Key
}

// NewKeyCache creates an empty KeyCache.
func NewKeyCache() *KeyCache {
	return &KeyCache{
		items: make(map[string]*models.Key),
	}
}

// Get returns the cached key for keyStr. The second return value is false when
// the key is not present in the cache.
func (c *KeyCache) Get(keyStr string) (*models.Key, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	k, ok := c.items[keyStr]
	return k, ok
}

// Set stores (or replaces) the key in the cache.
func (c *KeyCache) Set(keyStr string, k *models.Key) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[keyStr] = k
}

// Delete removes the key from the cache, forcing a fresh database read on the
// next access.
func (c *KeyCache) Delete(keyStr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, keyStr)
}
