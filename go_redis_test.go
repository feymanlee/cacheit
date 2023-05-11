package cacheit

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
)

var driverIndex = 0

func setupRedisDriver[V any](t *testing.T) *RedisDriver[V] {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to setup RedisDriver: %v", err)
		return nil
	}

	opt, err := redis.ParseURL("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("Failed to setup RedisDriver: %v", err)
		return nil
	}

	client := redis.NewClient(opt)
	driverIndex = driverIndex + 1
	driverName := "redis_test" + cast.ToString(driverIndex)
	rand.Seed(time.Now().UnixNano())
	prefix := "cache_prefix"
	if rand.Intn(2) > 0 {
		prefix = ""
	}
	err = RegisterRedisDriver(driverName, client, prefix)
	if err != nil {
		t.Fatalf("Failed register redisDriver: %v", err)
		return nil
	}
	driver, err := Use[V](driverName)
	if err != nil {
		t.Fatalf("Failed to setup RedisDriver: %v", err)
		return nil
	}
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
