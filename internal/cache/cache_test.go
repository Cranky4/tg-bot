package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRU_Set_WithKeyAndValue(t *testing.T) {
	cache := NewLRUCache(2)
	lruCache, ok := cache.(*LRUCache)
	assert.True(t, ok)
	cache.Set("default", "default")

	// добавляем 1й элемент
	cache.Set("key", "value")

	val, ex := lruCache.items["key"]
	assert.True(t, ex)

	lruItem, ok := val.Value.(*LRUCacheItem)
	assert.True(t, ok)

	assert.Equal(t, "value", lruItem.Value)

	// поиск несуществующего
	val, ex = lruCache.items["not"]
	assert.False(t, ex)

	// добавляем 2й элемент
	cache.Set("key2", "value2")

	val, ex = lruCache.items["key2"]
	assert.True(t, ex)
	lruItem, ok = val.Value.(*LRUCacheItem)
	assert.True(t, ok)

	assert.Equal(t, "value2", lruItem.Value)
	assert.True(t, ex)

	// 1й элемент в LRU кеше должен быть key2
	val = lruCache.lru.Front()
	lruItem, ok = val.Value.(*LRUCacheItem)
	assert.True(t, ok)
	assert.Equal(t, "key2", lruItem.Key)
	assert.Equal(t, "value2", lruItem.Value)

	// 1й вытеснился
	val, ex = lruCache.items["default"]
	assert.False(t, ex)
}

func TestLRU_Get_WithKey_value_hasValue(t *testing.T) {
	cache := NewLRUCache(2)
	lruCache, ok := cache.(*LRUCache)
	assert.True(t, ok)
	cache.Set("default", "default")

	// запрос существующего
	cache.Set("key", "value")
	val, ex := lruCache.Get("key")
	assert.True(t, ex)
	assert.Equal(t, "value", val)

	// запрос несуществующего
	val, ex = lruCache.Get("nothing")
	assert.Equal(t, nil, val)
	assert.False(t, ex)

	// вытеснение
	cache.Set("key2", "value2")
	val, ex = lruCache.Get("key2")
	assert.True(t, ex)
	assert.Equal(t, "value2", val)

	// 1й элемент в LRU кеше должен быть key2
	elem := lruCache.lru.Front()
	lruItem, ok := elem.Value.(*LRUCacheItem)
	assert.True(t, ok)
	assert.Equal(t, "key2", lruItem.Key)
	assert.Equal(t, "value2", lruItem.Value)

	val, ex = lruCache.Get("default")
	assert.Equal(t, nil, val)
	assert.False(t, ex)
}

func TestLRU_Len_notZero(t *testing.T) {
	cache := NewLRUCache(2)
	lru, ok := cache.(*LRUCache)
	assert.True(t, ok)
	cache.Set("default", "default")

	// добавляем 1й элемент
	lru.Set("key", "value")
	assert.Equal(t, 2, lru.Len())

	// добавляем 2й элемент, default вытесняется
	lru.Set("key2", "value2")
	assert.Equal(t, 2, lru.Len())
}

func TestLRU_Del_isTrue(t *testing.T) {
	cache := NewLRUCache(2)
	cache.Set("default", "default")

	lru, ok := cache.(*LRUCache)
	assert.True(t, ok)

	lru.Set("key", "value")

	// Удаляем сущесвтующий ключ
	assert.True(t, lru.Del("key"))

	// Удаляем несущесвтующий ключ
	assert.False(t, lru.Del("nothing"))
}
