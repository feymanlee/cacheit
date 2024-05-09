package cacheit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
)

// DriverType DriverType
type DriverType string

const (
	// driverRedis type redis
	driverRedis DriverType = "redis"
	// driverMemory type memory
	driverMemory DriverType = "memory"
)

// Many type many
type Many[V any] struct {
	Key   string
	Value V
	TTL   time.Duration
}

const (
	// NoExpirationTTL no expiration ttl
	NoExpirationTTL = time.Duration(-1)
	// ItemNotExistedTTL item not existed ttl
	ItemNotExistedTTL = time.Duration(-2)
)

var (
	ErrCacheMiss    = errors.New("cache not exists")
	ErrCacheExisted = errors.New("cache already existed")
)
var (
	defaultDriverName string
)

var registerDrivers sync.Map

// Driver cache driver interface
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
	// Del Remove an item from the cache alia Forget.
	Del(key string) error
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
	Remember(key string, ttl time.Duration, callback func() (V, error), force bool) (V, error)
	// RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
	RememberForever(key string, callback func() (V, error), force bool) (V, error)
	// RememberMany Get many item from the cache, or execute the given Closure and store the result.
	RememberMany(keys []string, ttl time.Duration, callback func(notHitKeys []string) (map[string]V, error), force bool) (map[string]V, error)
	// TTL Get cache ttl
	TTL(key string) (time.Duration, error)
	// WithCtx with context
	WithCtx(ctx context.Context) Driver[V]
	// WithSerializer with cache serializer
	WithSerializer(serializer Serializer) Driver[V]
}

type baseDriver struct {
	driverType  DriverType
	prefix      string
	redisClient *redis.Client
	memCache    *gocache.Cache
	serializer  Serializer
	// last error
	ctx context.Context
}

func newDriver(driverType DriverType, optionFns ...OptionFunc) (*baseDriver, error) {
	baseDriver := baseDriver{
		driverType: driverType,
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
	return &baseDriver, nil
}

// RegisterRedisDriver registers a Redis driver with the given driverName.
// This function creates a new driver based on the provided redis client and registers it in registerDrivers.
func RegisterRedisDriver(driverName string, redis *redis.Client, cacheKeyPrefix string) error {
	d, err := newDriver(driverRedis, withRedisClient(redis), withPrefix(cacheKeyPrefix))
	if err != nil {
		return err
	}
	_, loaded := registerDrivers.LoadOrStore(driverName, d)
	if loaded {
		return fmt.Errorf("redis driver: %s already registered", driverName)
	}
	return nil
}

// RegisterGoCacheDriver registers a GoCache driver with the given driverName.
// This function creates a new driver based on the provided go-cache client and registers it in registerDrivers.
func RegisterGoCacheDriver(driverName string, memCache *gocache.Cache, cacheKeyPrefix string) error {
	d, err := newDriver(driverMemory, withMemCache(memCache), withPrefix(cacheKeyPrefix))
	if err != nil {
		return err
	}
	_, loaded := registerDrivers.LoadOrStore(driverName, d)
	if loaded {
		return fmt.Errorf("go-cache driver: %s already registered", driverName)
	}
	return nil
}

// SetDefault set default driver
func SetDefault(driverName string) {
	defaultDriverName = driverName
}

// UnSetDefault cancel set default driver
func UnSetDefault() {
	defaultDriverName = ""
}

// UseDefault user default driver
func UseDefault[V any]() Driver[V] {
	d, err := Use[V](defaultDriverName)
	if err != nil {
		panic("default driver not set")
	}
	return d
}

// Use select a driver
func Use[V any](driverName string) (Driver[V], error) {
	if value, ok := registerDrivers.Load(driverName); ok {
		baseDriver := *value.(*baseDriver)
		switch baseDriver.driverType {
		case driverMemory:
			return &GoCacheDriver[V]{
				baseDriver,
			}, nil
		case driverRedis:
			return &RedisDriver[V]{
				baseDriver,
			}, nil
		default:
			return nil, fmt.Errorf("unsupport driver type: %s", baseDriver.driverType)
		}
	}
	return nil, fmt.Errorf("cached driver: %s not registered", driverName)
}

// get cache key
func (d *baseDriver) getCacheKey(key string) string {
	if d.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", d.prefix, key)
}
