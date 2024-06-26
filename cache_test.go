package cacheit

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

// TestRegisterRedisDriver tests the RegisterRedisDriver function
func TestRegisterRedisDriver(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	err := RegisterRedisDriver("redis_test", redisClient, "cache_prefix")
	assert.NoError(t, err)

	err = RegisterRedisDriver("redis_test", redisClient, "cache_prefix")
	assert.Error(t, err)

	SetDefault("redis_test")
}

// TestRegisterGoCacheDriver tests the RegisterGoCacheDriver function
func TestRegisterGoCacheDriver(t *testing.T) {
	memCache := cache.New(5*time.Minute, 10*time.Minute)
	err := RegisterGoCacheDriver("memory_test", memCache, "cache_prefix")
	assert.NoError(t, err)

	err = RegisterGoCacheDriver("memory_test", memCache, "cache_prefix")
	assert.Error(t, err)
}

// TestUse tests the Use function
func TestUse(t *testing.T) {
	d, err := Use[string]("redis_test")
	assert.NoError(t, err)
	assert.NotNil(t, d)

	d, err = Use[string]("memory_test")
	assert.NoError(t, err)
	assert.NotNil(t, d)

	d, err = Use[string]("non_existing_driver")
	assert.Nil(t, d)
	assert.Error(t, err)
}

// TestUseDefault tests the UseDefault function
func TestUseDefault(t *testing.T) {
	d := UseDefault[string]()
	assert.NotNil(t, d)
	assert.Panics(t, func() {
		UnSetDefault()
		_ = UseDefault[string]()
	}, "The function should panic when the default driver is not set")
}

func testCache[V any](t *testing.T, driver Driver[V], key string, value V) {
	duration := time.Second * 2
	var err error
	t.Run("add", func(t *testing.T) {
		err = driver.Add(key, value, duration)
		assert.NoError(t, err)

		err = driver.Add(key, value, duration)
		assert.ErrorIs(t, err, ErrCacheExisted)

		ttl, err := driver.TTL(key)
		assert.NoError(t, err)
		assert.True(t, ttl != 0)
		assert.LessOrEqual(t, ttl, duration)

		result, err := driver.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
		err = driver.Flush()
		assert.NoError(t, err)
	})
	t.Run("set and get and has", func(t *testing.T) {
		err = driver.Set(key, value, duration)
		assert.NoError(t, err)

		err = driver.Set(key, value, duration)
		assert.NoError(t, err)
		has, err := driver.Has(key)
		assert.NoError(t, err)
		assert.True(t, has)

		result, err := driver.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
		err = driver.Flush()
		assert.NoError(t, err)

		has, err = driver.Has(key)
		assert.NoError(t, err)
		assert.True(t, !has)
	})
	t.Run("remember", func(t *testing.T) {
		result, err := driver.Remember(key, duration, func() (V, error) {
			return value, nil
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
		err = driver.Flush()
		assert.NoError(t, err)
	})
	t.Run("remember forever", func(t *testing.T) {
		result, err := driver.RememberForever(key, func() (V, error) {
			return value, nil
		}, false)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		ttl, err := driver.TTL(key)
		assert.NoError(t, err)
		assert.Equal(t, NoExpirationTTL, ttl)

		err = driver.Flush()
		assert.NoError(t, err)
	})
	t.Run("forever", func(t *testing.T) {
		err = driver.Forever(key, value)
		assert.NoError(t, err)

		result, err := driver.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, result, value)

		ttl, err := driver.TTL(key)
		assert.NoError(t, err)
		assert.Equal(t, NoExpirationTTL, ttl)

		err = driver.Flush()
		assert.NoError(t, err)
	})
	t.Run("forget", func(t *testing.T) {
		err = driver.Forever(key, value)
		assert.NoError(t, err)

		err = driver.Forget(key)
		assert.NoError(t, err)

		result, err := driver.Get(key)
		assert.Zerof(t, result, "")
		assert.ErrorIs(t, err, ErrCacheMiss)

		ttl, _ := driver.TTL(key)
		assert.Equal(t, ItemNotExistedTTL, ttl)

		has, _ := driver.Has(key)
		assert.True(t, !has)
	})
	t.Run("del", func(t *testing.T) {
		err = driver.Forever(key, value)
		assert.NoError(t, err)

		err = driver.Del(key)
		assert.NoError(t, err)

		result, err := driver.Get(key)
		assert.Zerof(t, result, "")
		assert.ErrorIs(t, err, ErrCacheMiss)

		ttl, _ := driver.TTL(key)
		assert.Equal(t, ItemNotExistedTTL, ttl)

		has, _ := driver.Has(key)
		assert.True(t, !has)
	})

	t.Run("ttl", func(t *testing.T) {
		err = driver.Forever(key, value)
		assert.NoError(t, err)

		err = driver.Del(key)
		assert.NoError(t, err)

		result, err := driver.Get(key)
		assert.Zerof(t, result, "")
		assert.ErrorIs(t, err, ErrCacheMiss)

		ttl, _ := driver.TTL(key)
		assert.Equal(t, ItemNotExistedTTL, ttl)

		has, _ := driver.Has(key)
		assert.True(t, !has)
	})

	t.Run("many and set many", func(t *testing.T) {
		expected := make(map[string]V)
		var items []Many[V]
		itemsCount := 10
		kes := make([]string, 0, itemsCount)
		for i := 0; i < itemsCount; i++ {
			k := key + cast.ToString(i)
			expected[k] = value
			kes = append(kes, k)
			items = append(items, Many[V]{
				Key:   k,
				Value: value,
				TTL:   duration,
			})
		}
		err = driver.SetMany(items)
		assert.NoError(t, err)

		result, err := driver.Many(kes)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)

		ttl, err := driver.TTL(kes[3])
		assert.NoError(t, err)
		assert.True(t, ttl != 0)
		assert.LessOrEqual(t, ttl, duration)

		has, _ := driver.Has(key)
		assert.True(t, !has)
	})
}

func testNumberCache[V any](t *testing.T, driver Driver[V], key string, value V) {
	duration := time.Second * 10
	var err error
	t.Run("set number", func(t *testing.T) {
		if isNumeric(value) {
			err = driver.SetNumber(key, value, duration)
			assert.NoError(t, err)

			_, err = driver.Increment(key, value)
			assert.NoError(t, err)

			value2, err := driver.Decrement(key, value)
			assert.NoError(t, err)
			assert.Equal(t, value, value2)

			err = driver.Flush()
			assert.NoError(t, err)
		} else {
			err = driver.SetNumber(key, value, duration)
			assert.Error(t, err)

			_, err = driver.Increment(key, value)
			assert.Error(t, err)

			_, err = driver.Decrement(key, value)
			assert.Error(t, err)
		}
	})
}
