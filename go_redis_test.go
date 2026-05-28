package cacheit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var driverIndex uint64

func nextDriverName(prefix string) string {
	return prefix + "_" + cast.ToString(atomic.AddUint64(&driverIndex, 1))
}

func setupRedisDriver[V any](t *testing.T) *RedisDriver[V] {
	t.Helper()
	return setupRedisDriverWithPrefix[V](t, "cache_prefix")
}

func setupRedisDriverWithPrefix[V any](t *testing.T, prefix string) *RedisDriver[V] {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err, "setup miniredis")
	t.Cleanup(mr.Close)

	opt, err := redis.ParseURL("redis://" + mr.Addr())
	require.NoError(t, err, "parse miniredis URL")

	client := redis.NewClient(opt)
	t.Cleanup(func() {
		require.NoError(t, client.Close(), "close redis client")
	})

	driverName := nextDriverName("redis_test")
	err = RegisterRedisDriver(driverName, client, prefix)
	require.NoError(t, err, "register redis driver")

	driver, err := Use[V](driverName)
	require.NoError(t, err, "use redis driver")

	driver.WithCtx(context.Background())
	driver.WithSerializer(&JSONSerializer{})
	return driver.(*RedisDriver[V])
}

func TestRedisDriver(t *testing.T) {
	redisDriverString := setupRedisDriver[string](t)
	testCache[string](t, redisDriverString, "test_string_key", "test_string_value")

	redisDriverStruct := setupRedisDriver[testStruct](t)
	testCache[testStruct](t, redisDriverStruct, "test_struct_key", testStructData)

	testNumberCache[string](t, redisDriverString, "test_string_key", "test_string_value")

	redisDriverInt := setupRedisDriver[int](t)
	testNumberCache[int](t, redisDriverInt, "test_int_key", 2)

	redisDriverUint := setupRedisDriver[uint](t)
	testNumberCache[uint](t, redisDriverUint, "test_uint_key", uint(2))

	redisDriverFloat := setupRedisDriver[float32](t)
	testNumberCache[float32](t, redisDriverFloat, "test_float_key", float32(2.0))
}

func TestRedisForeverOverwritesExistingTTLWithNoExpiration(t *testing.T) {
	driver := setupRedisDriver[string](t)
	key := "forever_overwrites_ttl"

	assert.NoError(t, driver.Set(key, "temporary", time.Minute))
	assert.NoError(t, driver.Forever(key, "permanent"))

	ttl, err := driver.TTL(key)
	assert.NoError(t, err)
	assert.Equal(t, NoExpirationTTL, ttl)

	got, err := driver.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, "permanent", got)
}

func TestRedisRememberForeverWritesNoExpiration(t *testing.T) {
	driver := setupRedisDriver[string](t)
	key := "remember_forever_no_expiration"

	got, err := driver.RememberForever(key, func() (string, error) {
		return "permanent", nil
	}, true)
	assert.NoError(t, err)
	assert.Equal(t, "permanent", got)

	ttl, err := driver.TTL(key)
	assert.NoError(t, err)
	assert.Equal(t, NoExpirationTTL, ttl)
}

func TestRedisSetManySupportsExpirationModes(t *testing.T) {
	driver := setupRedisDriver[string](t)

	err := driver.SetMany([]Many[string]{
		{Key: "set_many_no_expiration_zero", Value: "zero", TTL: 0},
		{Key: "set_many_no_expiration_constant", Value: "constant", TTL: NoExpirationTTL},
		{Key: "set_many_expiring", Value: "ttl", TTL: time.Minute},
	})
	assert.NoError(t, err)

	zeroTTL, err := driver.TTL("set_many_no_expiration_zero")
	assert.NoError(t, err)
	assert.Equal(t, NoExpirationTTL, zeroTTL)

	constantTTL, err := driver.TTL("set_many_no_expiration_constant")
	assert.NoError(t, err)
	assert.Equal(t, NoExpirationTTL, constantTTL)

	expiringTTL, err := driver.TTL("set_many_expiring")
	assert.NoError(t, err)
	assert.True(t, expiringTTL > 0)
	assert.LessOrEqual(t, expiringTTL, time.Minute)
}

func TestRedisManyAndDelManyEmptyInput(t *testing.T) {
	driver := setupRedisDriver[string](t)

	got, err := driver.Many(nil)
	assert.NoError(t, err)
	assert.Empty(t, got)

	assert.NoError(t, driver.DelMany(nil))
}

func TestRedisFlushWithPrefixOnlyDeletesMatchingKeys(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	opt, err := redis.ParseURL("redis://" + mr.Addr())
	require.NoError(t, err)
	client := redis.NewClient(opt)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	driverNameA := nextDriverName("redis_prefix_a")
	driverNameB := nextDriverName("redis_prefix_b")
	require.NoError(t, RegisterRedisDriver(driverNameA, client, "prefix_a"))
	require.NoError(t, RegisterRedisDriver(driverNameB, client, "prefix_b"))

	driverA, err := Use[string](driverNameA)
	require.NoError(t, err)
	driverB, err := Use[string](driverNameB)
	require.NoError(t, err)

	assert.NoError(t, driverA.Set("shared", "a", time.Minute))
	assert.NoError(t, driverB.Set("shared", "b", time.Minute))
	assert.NoError(t, client.Set(context.Background(), "unprefixed", "raw", time.Minute).Err())

	assert.NoError(t, driverA.Flush())

	_, err = driverA.Get("shared")
	assert.ErrorIs(t, err, ErrCacheMiss)

	gotB, err := driverB.Get("shared")
	assert.NoError(t, err)
	assert.Equal(t, "b", gotB)

	raw, err := client.Get(context.Background(), "unprefixed").Result()
	assert.NoError(t, err)
	assert.Equal(t, "raw", raw)
}

func TestRedisFlushWithoutPrefixDeletesAllKeys(t *testing.T) {
	driver := setupRedisDriverWithPrefix[string](t, "")

	assert.NoError(t, driver.Set("managed", "value", time.Minute))
	assert.NoError(t, driver.redisClient.Set(context.Background(), "raw", "value", time.Minute).Err())

	assert.NoError(t, driver.Flush())

	_, err := driver.Get("managed")
	assert.ErrorIs(t, err, ErrCacheMiss)

	_, err = driver.redisClient.Get(context.Background(), "raw").Result()
	assert.ErrorIs(t, err, redis.Nil)
}

func TestDefaultDriverNameConcurrentAccess(t *testing.T) {
	driver := setupRedisDriver[string](t)
	driverName := nextDriverName("default_driver_race")
	require.NoError(t, RegisterRedisDriver(driverName, driver.redisClient, "default_driver_race"))

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				SetDefault(driverName)
				func() {
					defer func() {
						_ = recover()
					}()
					_ = UseDefault[string]()
				}()
				UnSetDefault()
			}
		}()
	}
	wg.Wait()
}
