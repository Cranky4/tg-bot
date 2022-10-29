package memory

import (
	"container/list"
	"sync"

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

func (c *LRUCache) Set(key string, value any) {
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
}

func (c *LRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.exists(key) {
		return nil, false
	}

	val := c.items[key]
	c.lru.MoveToFront(val)

	item, ok := val.Value.(*LRUCacheItem)
	if ok {
		return item.Value, true
	}

	return nil, false
}

func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lru.Len()
}

func (c *LRUCache) Del(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.exists(key) {
		return false
	}

	val := c.items[key]

	item, ok := c.lru.Remove(val).(*LRUCacheItem)
	if ok {
		delete(c.items, item.Key)
		return true
	}

	return false
}

func (c *LRUCache) exists(key string) bool {
	_, ok := c.items[key]

	return ok
}
