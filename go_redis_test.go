package cacheit

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

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
	driver, err := New[V](DriverRedis,
		WithRedisClient(client),
		WithPrefix("a"),
		WithSerializer(&JSONSerializer{}),
	)
	driver = driver.WithCtx(context.Background())
	if err != nil {
		t.Fatalf("Failed to setup RedisDriver: %v", err)
		return nil
	}

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
