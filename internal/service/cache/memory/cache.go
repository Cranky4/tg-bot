package memory

import (
	"container/list"
	"context"
	"sync"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache"
)

type LRUCache struct {
	mu    *sync.RWMutex
	len   int
	items map[string]*list.Element
	lru   *list.List
}

type LRUCacheItem struct {
	Key   string
	Value any
}

func NewLRUCache(len int) cache.Cache {
	return &LRUCache{
		mu:    &sync.RWMutex{},
		items: make(map[string]*list.Element, len),
		lru:   list.New(), // []*LRUCacheItem
		len:   len,
	}
}

func (c *LRUCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.exists(key) && c.lru.Len() == c.len {
		if last := c.lru.Back(); last != nil {
			item, ok := last.Value.(*LRUCacheItem)
			if ok {
				delete(c.items, item.Key)
			}
			c.lru.Remove(last)
		}
	}

	elem := c.lru.PushFront(&LRUCacheItem{Key: key, Value: value})

	c.items[key] = elem

	return nil
}

func (c *LRUCache) Get(ctx context.Context, key string) (any, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.exists(key) {
		return nil, false, nil
	}

	val := c.items[key]
	c.lru.MoveToFront(val)

	item, ok := val.Value.(*LRUCacheItem)
	if ok {
		return item.Value, true, nil
	}

	return nil, false, nil
}

func (c *LRUCache) Len() (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lru.Len(), nil
}

func (c *LRUCache) Del(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.exists(key) {
		return false, nil
	}

	val := c.items[key]

	item, ok := c.lru.Remove(val).(*LRUCacheItem)
	if ok {
		delete(c.items, item.Key)
		return true, nil
	}

	return false, nil
}

func (c *LRUCache) exists(key string) bool {
	_, ok := c.items[key]

	return ok
}
