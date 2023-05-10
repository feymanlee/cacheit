package cacheit

import (
	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
)

type OptionFunc func(driver *baseDriver) error

//
// WithSerializer
//  @Description: set a cache serializer
//  @param serializer
//  @return OptionFunc
//
func WithSerializer(serializer Serializer) OptionFunc {
	return func(driver *baseDriver) error {
		driver.serializer = serializer
		return nil
	}
}

//
// WithPrefix
//  @Description:
//  @param prefix
//  @return OptionFunc
//
func WithPrefix(prefix string) OptionFunc {
	return func(driver *baseDriver) error {
		driver.prefix = prefix
		return nil
	}
}

//
// WithRedisClient
//  @Description:
//  @param redis
//  @return OptionFunc
//
func WithRedisClient(redis *redis.Client) OptionFunc {
	return func(driver *baseDriver) error {
		driver.redisClient = redis
		return nil
	}
}

//
// WithMemCache
//  @Description:
//  @param memCache
//  @return OptionFunc
//
func WithMemCache(memCache *gocache.Cache) OptionFunc {
	return func(driver *baseDriver) error {
		driver.memCache = memCache
		return nil
	}
}
