package cache

import (
	"context"
	"time"
)

const (
	MemoryMode = "memory"
	RedisMode  = "redis"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (any, bool, error)
	Del(ctx context.Context, key string) (bool, error)
}
