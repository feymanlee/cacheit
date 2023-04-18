/**
 * @Author: lifameng@changba.com
 * @Description:
 * @File:  cache
 * @Date: 2023/4/13 15:06
 */

package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
)

type DriverType string

const (
	DriverRedis  DriverType = "redis"
	DriverMemory DriverType = "memory"
)

type Many[V any] struct {
	Key   string
	Value V
	TTL   time.Duration
}

type Driver[V any] interface {
	// Add Store an item in the cache if the key doesn't exist.
	Add(key string, value V, t time.Duration) error
	// Put Store an item in the cache for a given number of seconds.
	Put(key string, value V, t time.Duration) error
	// PutMany Store multiple items in the cache for a given number of seconds.
	PutMany(many []Many[V]) error
	// Forever Store an item in the cache indefinitely.
	Forever(key string, value V) error
	// Forget Remove an item from the cache.
	Forget(key string) error
	// Flush Remove all items from the cache.
	Flush() error
	// Get Retrieve an item from the cache by key.
	Get(key string) (V, error)
	// Has Determine if an item exists in the cache.
	Has(key string) (bool, error)
	// Many Retrieve multiple items from the cache by key.
	// Items not found in the cache will have a nil value.
	Many(keys []string) (map[string]V, error)
	// SetInt64 set the int64 value of an item in the cache.
	SetInt64(key string, value int64, t time.Duration) error
	// IncrementInt64 Increment the value of an item in the cache.
	IncrementInt64(key string, value int64) (int64, error)
	// DecrementInt64 Decrement the value of an item in the cache.
	DecrementInt64(key string, value int64) (int64, error)
	// Remember Get an item from the cache, or execute the given Closure and store the result.
	Remember(key string, ttl time.Duration, callback func() (V, error)) (V, error)
	// RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
	RememberForever(key string, callback func() (V, error)) (V, error)
}

type baseDriver struct {
	prefix      string
	redisClient *redis.Client
	memCache    *gocache.Cache
	serializer  Serializer
	// last error
	ctx context.Context
}

func New[V any](driver DriverType, optionFns ...OptionFunc) (Driver[V], error) {
	baseDriver := baseDriver{
		ctx:        context.Background(),
		serializer: &JSONSerializer{},
	}
	if len(optionFns) > 0 {
		for _, optionFn := range optionFns {
			err := optionFn(&baseDriver)
			if err != nil {
				return nil, err
			}
		}
	}
	switch driver {
	case DriverRedis:
		if baseDriver.redisClient == nil {
			return nil, errors.New("redis client is not initialized")
		}
		return &RedisDriver[V]{
			baseDriver,
		}, nil
	case DriverMemory:
		if baseDriver.memCache == nil {
			return nil, errors.New("go-cache client is not initialized")
		}
		return &GoCacheDriver[V]{
			baseDriver,
		}, nil
	default:
		return nil, errors.New("unsupported driver")
	}
}

func (d *baseDriver) getCacheKey(key string) string {
	if d.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", d.prefix, key)
}
