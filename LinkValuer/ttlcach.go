package linkvaluer

import (
	"sync"
	"time"
)

type item[V any] struct {
	value  V
	expiry time.Time
}

func (i item[V]) isExpired() bool { return time.Now().After(i.expiry) }

type TTLCache[K comparable, V any] struct {
	items map[K]item[V]
	mu    sync.Mutex
}

func NewTTL[K comparable, V any](ttl time.Duration) *TTLCache[K, V] {
	c := &TTLCache[K, V]{items: make(map[K]item[V])}
	go func() {
		for range time.Tick(ttl) {
			c.mu.Lock()
			for k, it := range c.items {
				if it.isExpired() {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}()
	return c
}

func (c *TTLCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = item[V]{value: value, expiry: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	it, ok := c.items[key]
	if ok && it.isExpired() {
		delete(c.items, key)
		ok = false
	}
	c.mu.Unlock()
	return it.value, ok
}

func (c *TTLCache[K, V]) Remove(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
