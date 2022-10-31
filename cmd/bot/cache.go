package main

import (
	"fmt"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache"
	memory_cache "gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache/memory"
	redis_cache "gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache/redis"
)

const undefinedMode = "неизвестный режим кеширования: %s"

func initCache(conf config.Config) (cache.Cache, error) {
	switch conf.Cache.Mode {
	case cache.MemoryMode:
		return memory_cache.NewLRUCache(conf.Cache.Length), nil
	case cache.RedisMode:
		return redis_cache.NewRedisCache(conf.Redis), nil
	default:
		return nil, fmt.Errorf(undefinedMode, conf.Cache.Mode)
	}
}
