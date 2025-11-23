package cache

import (
	"sync"
	"time"
)

type Item struct {
	Value      interface{}
	Expiration int64
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]*Item
	ttl   time.Duration
}

func New(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]*Item),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go c.cleanupLoop()

	return c
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(c.ttl).UnixNano()
	c.items[key] = &Item{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for key, item := range c.items {
		if now > item.Expiration {
			delete(c.items, key)
		}
	}
}

func MakeKey(name, tag string) string {
	return name + "#" + tag
}
