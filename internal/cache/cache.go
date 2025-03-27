package cache

import (
	"sync"
	"time"
)

// Item represents a cached item
type Item struct {
	Value      interface{}
	Expiration int64
}

// Expired returns true if the item has expired
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// Cache is a simple in-memory cache
type Cache struct {
	items map[string]Item
	mu    sync.RWMutex
}

// New creates a new cache instance with cleanup
func New(cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]Item),
	}

	// Start cleanup goroutine if needed
	if cleanupInterval > 0 {
		go cache.cleanupLoop(cleanupInterval)
	}

	return cache
}

// Set adds an item to the cache with optional expiration
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if item.Expired() {
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]Item)
}

// cleanup removes expired items from the cache
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.Expired() {
			delete(c.items, key)
		}
	}
}

// cleanupLoop runs cleanup at the specified interval
func (c *Cache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.cleanup()
	}
}
