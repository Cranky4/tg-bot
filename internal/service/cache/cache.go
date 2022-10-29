package cache

const (
	MemoryMode = "memory"
	RedisMode  = "redis"
)

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Len() int
	Del(key string) bool
}
