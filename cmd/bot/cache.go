package main

import (
	"fmt"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache"
	memory_cache "gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache/memory"
)

const undefinedMode = "неизвестный режим кеширования: %s"

func initCache(conf config.CacheConf) (cache.Cache, error) {
	switch conf.Mode {
	case cache.MemoryMode:
		return memory_cache.NewLRUCache(conf.Length), nil
	case cache.RedisMode:
		return nil, fmt.Errorf("implement me!")
	default:
		return nil, fmt.Errorf(undefinedMode, conf.Mode)
	}
}
