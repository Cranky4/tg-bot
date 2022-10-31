package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache"
)

const (
	setErrorMsg = "redis set error"
	getErrorMsg = "redis get error"
	delErrorMsg = "redis del error"
)

type redisCache struct {
	rdb *redis.Client
}

func NewRedisCache(config config.RedisConf) cache.Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &redisCache{rdb: rdb}
}

func (r *redisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := r.rdb.Set(ctx, key, value, expiration).Err(); err != nil {
		return errors.Wrap(err, setErrorMsg)
	}

	return nil
}

func (r *redisCache) Get(ctx context.Context, key string) (any, bool, error) {
	cmd := r.rdb.Get(ctx, key)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, errors.Wrap(err, getErrorMsg)
	}

	value, _ := cmd.Result()

	return value, true, nil
}

func (r *redisCache) Del(ctx context.Context, key string) (bool, error) {
	cmd := r.rdb.Del(ctx, key)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, errors.Wrap(err, delErrorMsg)
	}

	return true, nil
}
