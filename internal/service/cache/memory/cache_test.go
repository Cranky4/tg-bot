package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRU_Set_WithKeyAndValue(t *testing.T) {
	cache := NewLRUCache(2)
	lruCache, ok := cache.(*LRUCache)
	assert.True(t, ok)
	ctx := context.Background()

	cache.Set(ctx, "default", "default", 1*time.Minute)

	// добавляем 1й элемент
	cache.Set(ctx, "key", "value", 1*time.Minute)

	val, ex := lruCache.items["key"]
	assert.True(t, ex)

	lruItem, ok := val.Value.(*LRUCacheItem)
	assert.True(t, ok)

	assert.Equal(t, "value", lruItem.Value)

	// поиск несуществующего
	val, ex = lruCache.items["not"]
	assert.False(t, ex)

	// добавляем 2й элемент
	cache.Set(ctx, "key2", "value2", 1*time.Minute)

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
	ctx := context.Background()
	cache.Set(ctx, "default", "default", 1*time.Minute)

	// запрос существующего
	cache.Set(ctx, "key", "value", 1*time.Minute)
	val, ex, err := lruCache.Get(ctx, "key")
	assert.Nil(t, err)
	assert.True(t, ex)
	assert.Equal(t, "value", val)

	// запрос несуществующего
	val, ex, err = lruCache.Get(ctx, "nothing")
	assert.Nil(t, err)
	assert.Equal(t, nil, val)
	assert.False(t, ex)

	// вытеснение
	cache.Set(ctx, "key2", "value2", 1*time.Minute)
	val, ex, err = lruCache.Get(ctx, "key2")
	assert.Nil(t, err)
	assert.True(t, ex)
	assert.Equal(t, "value2", val)

	// 1й элемент в LRU кеше должен быть key2
	elem := lruCache.lru.Front()
	lruItem, ok := elem.Value.(*LRUCacheItem)
	assert.True(t, ok)
	assert.Equal(t, "key2", lruItem.Key)
	assert.Equal(t, "value2", lruItem.Value)

	val, ex, err = lruCache.Get(ctx, "default")
	assert.Nil(t, err)
	assert.Equal(t, nil, val)
	assert.False(t, ex)
}

func TestLRU_Len_notZero(t *testing.T) {
	cache := NewLRUCache(2)
	lru, ok := cache.(*LRUCache)
	assert.True(t, ok)
	ctx := context.Background()
	cache.Set(ctx, "default", "default", 1*time.Minute)

	// добавляем 1й элемент
	lru.Set(ctx, "key", "value", 1*time.Minute)
	l, err := lru.Len()
	assert.Nil(t, err)
	assert.Equal(t, 2, l)

	// добавляем 2й элемент, default вытесняется
	lru.Set(ctx, "key2", "value2", 1*time.Minute)
	l, err = lru.Len()
	assert.Nil(t, err)
	assert.Equal(t, 2, l)
}

func TestLRU_Del_isTrue(t *testing.T) {
	cache := NewLRUCache(2)
	ctx := context.Background()
	cache.Set(ctx, "default", "default", 1*time.Minute)

	lru, ok := cache.(*LRUCache)
	assert.True(t, ok)

	lru.Set(ctx, "key", "value", 1*time.Minute)

	// Удаляем сущесвтующий ключ
	ok, err := lru.Del(ctx, "key")
	assert.Nil(t, err)
	assert.True(t, ok)

	// Удаляем несущесвтующий ключ
	ok, err = lru.Del(ctx, "nothing")
	assert.Nil(t, err)
	assert.False(t, ok)
}
