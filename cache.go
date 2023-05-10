package cacheit

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

const (
	NoExpirationTTL   = time.Duration(-1)
	ItemNotExistedTTL = time.Duration(-2)
)

var (
	CacheMissError    = errors.New("cache not exists")
	CacheExistedError = errors.New("cache already existed")
)

type Driver[V any] interface {
	// Add Store an item in the cache if the key doesn't exist.
	Add(key string, value V, t time.Duration) error
	// Set Store an item in the cache for a given number of seconds.
	Set(key string, value V, t time.Duration) error
	// SetMany Store multiple items in the cache for a given number of seconds.
	SetMany(many []Many[V]) error
	// Forever Store an item in the cache indefinitely.
	Forever(key string, value V) error
	// Forget Remove an item from the cache.
	Forget(key string) error
	// Flush Remove all items from the cache.
	Flush() error
	// Get Retrieve an item from the cache by key.
	Get(key string) (V, error)
	// Has Determined if an item exists in the cache.
	Has(key string) (bool, error)
	// Many Retrieve multiple items from the cache by key.
	// Items not found in the cache will have a nil value.
	Many(keys []string) (map[string]V, error)
	// SetNumber set the int64 value of an item in the cache.
	SetNumber(key string, value V, t time.Duration) error
	// Increment the value of an item in the cache.
	Increment(key string, n V) (V, error)
	// Decrement the value of an item in the cache.
	Decrement(key string, n V) (V, error)
	// Remember Get an item from the cache, or execute the given Closure and store the result.
	Remember(key string, ttl time.Duration, callback func() (V, error)) (V, error)
	// RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
	RememberForever(key string, callback func() (V, error)) (V, error)
	// TTL Get cache ttl
	TTL(key string) (time.Duration, error)
	// WithCtx with context
	WithCtx(ctx context.Context) Driver[V]
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
			if err := optionFn(&baseDriver); err != nil {
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
