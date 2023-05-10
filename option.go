package cacheit

import (
	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
)

type OptionFunc func(driver *baseDriver) error

// WithSerializer set a cache serializer
func WithSerializer(serializer Serializer) OptionFunc {
	return func(driver *baseDriver) error {
		driver.serializer = serializer
		return nil
	}
}

// WithPrefix  set a cache prefix
func WithPrefix(prefix string) OptionFunc {
	return func(driver *baseDriver) error {
		driver.prefix = prefix
		return nil
	}
}

// WithRedisClient set a redis client
func WithRedisClient(redis *redis.Client) OptionFunc {
	return func(driver *baseDriver) error {
		driver.redisClient = redis
		return nil
	}
}

// WithMemCache with a go-cache client
func WithMemCache(memCache *gocache.Cache) OptionFunc {
	return func(driver *baseDriver) error {
		driver.memCache = memCache
		return nil
	}
}
