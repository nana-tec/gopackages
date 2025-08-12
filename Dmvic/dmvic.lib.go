package dmvic

import (
	"sync"
	"time"
)

// Package dmvic provides a generic cache implementation with support for time-to-live (TTL) expiration.

// item represents a cache item with a value and an expiration time.
// It is used internally by TTLCache to store values with their expiration timestamps.
type item[V any] struct {
	value  V         // The cached value
	expiry time.Time // Expiration timestamp for this item
}

// isExpired checks if the cache item has expired.
// Returns true if the current time is after the item's expiry time.
func (i item[V]) isExpired() bool {
	return time.Now().After(i.expiry)
}

// DmvitokenStorage defines the interface for token storage operations.
// It provides methods for storing, retrieving, and managing tokens with TTL functionality.
type DmvitokenStorage interface {
	// Set stores a token with the specified key, value, and time-to-live duration.
	Set(key string, value string, ttl time.Duration)

	// Get retrieves a token by key, returning the value and a boolean indicating if found.
	Get(key string) (string, bool)

	// Remove deletes a token by key from storage.
	Remove(key string)

	// Pop retrieves and removes a token by key, returning the value and a boolean indicating if found.
	Pop(key string) (string, bool)
}

// TTLCache is a generic cache implementation with support for time-to-live (TTL) expiration.
// It provides thread-safe operations for storing and retrieving items with automatic cleanup
// of expired entries.
type TTLCache[K comparable, V any] struct {
	items map[K]item[V] // The map storing cache items
	mu    sync.Mutex    // Mutex for controlling concurrent access to the cache
}

// NewTTL creates a new TTLCache instance and starts a goroutine to periodically
// remove expired items. The cleanup interval is set to the provided TTL duration.
// Returns a pointer to the new TTLCache instance.
func NewTTL[K comparable, V any](ttl time.Duration) *TTLCache[K, V] {
	c := &TTLCache[K, V]{
		items: make(map[K]item[V]),
	}

	go func() {
		// 5  * time.Second  5 sec

		for range time.Tick(ttl) {
			c.mu.Lock()

			// Iterate over the cache items and delete expired ones.
			for key, item := range c.items {
				if item.isExpired() {
					delete(c.items, key)
				}
			}

			c.mu.Unlock()
		}
	}()

	return c
}

// Set adds a new item to the cache with the specified key, value, and time-to-live (TTL).
// If an item with the same key already exists, it will be overwritten with the new value and TTL.
// This operation is thread-safe.
func (c *TTLCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item[V]{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

// Get retrieves the value associated with the given key from the cache.
// Returns the value and true if found and not expired, or the zero value and false otherwise.
// This operation is thread-safe and automatically removes expired items when accessed.
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		// If the key is not found, return the zero value for V and false.
		return item.value, false
	}

	if item.isExpired() {
		// If the item has expired, remove it from the cache and return the
		// value and false.
		delete(c.items, key)
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}

// Remove removes the item with the specified key from the cache.
// This operation is thread-safe and does not return any value.
func (c *TTLCache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Delete the item with the given key from the cache.
	delete(c.items, key)
}

// Pop removes and returns the item with the specified key from the cache.
// Returns the value and true if the item exists and is not expired,
// or the zero value and false otherwise. This operation is thread-safe.
func (c *TTLCache[K, V]) Pop(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		// If the key is not found, return the zero value for V and false.
		return item.value, false
	}

	// If the key is found, delete the item from the cache.
	delete(c.items, key)

	if item.isExpired() {
		// If the item has expired, return the value and false.
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}
