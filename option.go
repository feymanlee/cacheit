package cacheit

import (
	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
)

type OptionFunc func(driver *baseDriver) error

// withPrefix  set a cache prefix
func withPrefix(prefix string) OptionFunc {
	return func(driver *baseDriver) error {
		driver.prefix = prefix
		return nil
	}
}

// withRedisClient set a redis client
func withRedisClient(redis *redis.Client) OptionFunc {
	return func(driver *baseDriver) error {
		driver.redisClient = redis
		return nil
	}
}

// withMemCache with a go-cache client
func withMemCache(memCache *gocache.Cache) OptionFunc {
	return func(driver *baseDriver) error {
		driver.memCache = memCache
		return nil
	}
}
