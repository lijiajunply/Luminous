package util

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/singleflight"
)

var (
	cacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "luminous_cache_hits_total",
		Help: "Total cache hits.",
	}, []string{"key"})

	cacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "luminous_cache_misses_total",
		Help: "Total cache misses.",
	}, []string{"key"})
)

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

const defaultMaxEntries = 10000

type Cache struct {
	mu         sync.RWMutex
	items      map[string]*cacheItem
	stopCh     chan struct{}
	sf         singleflight.Group
	maxEntries int
	name       string
}

func NewCache() *Cache {
	return NewCacheWithName("default")
}

func NewCacheWithName(name string) *Cache {
	c := &Cache{
		items:      make(map[string]*cacheItem),
		stopCh:     make(chan struct{}),
		maxEntries: defaultMaxEntries,
		name:       name,
	}
	go c.cleanupLoop()
	return c
}

func (c *Cache) Stop() {
	close(c.stopCh)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	if ok && time.Now().Before(item.expiration) {
		val := item.value
		c.mu.RUnlock()
		cacheHits.WithLabelValues(c.name).Inc()
		return val, true
	}
	c.mu.RUnlock()

	cacheMisses.WithLabelValues(c.name).Inc()
	if ok {
		c.mu.Lock()
		if current, exists := c.items[key]; exists && current == item {
			delete(c.items, key)
		}
		c.mu.Unlock()
	}
	return nil, false
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	if len(c.items) >= c.maxEntries {
		c.evictLocked(c.maxEntries / 10)
	}
	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

func (c *Cache) evictLocked(n int) {
	for k := range c.items {
		if n <= 0 {
			break
		}
		delete(c.items, k)
		n--
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache) GetOrSet(key string, ttl time.Duration, factory func() (interface{}, error)) (interface{}, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}
	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		val, err := factory()
		if err != nil {
			return nil, err
		}
		c.Set(key, val, ttl)
		return val, nil
	})
	return v, err
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.items {
				if now.After(v.expiration) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}
}
