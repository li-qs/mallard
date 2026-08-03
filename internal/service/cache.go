package service

import (
	"sync"
	"time"
)

type cacheEntry[V any] struct {
	value    V
	expireAt time.Time
}

type ttlCache[K comparable, V any] struct {
	mu   sync.Mutex
	ttl  time.Duration
	data map[K]cacheEntry[V]
}

func newTTLCache[K comparable, V any](ttl time.Duration) *ttlCache[K, V] {
	return &ttlCache[K, V]{
		ttl:  ttl,
		data: make(map[K]cacheEntry[V]),
	}
}

func (c *ttlCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.data[key]
	if !ok {
		var zero V
		return zero, false
	}
	if time.Now().After(e.expireAt) {
		delete(c.data, key)
		var zero V
		return zero, false
	}
	return e.value, true
}

func (c *ttlCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry[V]{
		value:    value,
		expireAt: time.Now().Add(c.ttl),
	}
}

func (c *ttlCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}
