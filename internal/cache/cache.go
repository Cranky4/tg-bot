package cache

import (
	"container/list"
	"sync"
)

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Len() int
	Del(key string) bool
}

type LRUCache struct {
	mu    *sync.RWMutex
	len   int
	items map[string]any
	lru   *list.List
}

func NewLRUCache() Cache {
	return &LRUCache{
		mu:    &sync.RWMutex{},
		items: make(map[string]any),
		lru:   list.New(),
	}
}

func (c *LRUCache) Get(key string) (any, bool) {
	val, ok := c.items[key]

	return val, ok
}

func (c *LRUCache) Set(key string, value any) {
	c.items[key] = value
	c.lru.PushFront(key)
}

func (c *LRUCache) Len() int {
	return c.len
}

func (c *LRUCache) Del(key string) bool {
	if _, ok := c.items[key]; !ok {
		return false
	}

	delete(c.items, key)

	return true
}
