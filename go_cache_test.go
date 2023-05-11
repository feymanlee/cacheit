package cacheit

import (
	"context"
	"math/rand"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/spf13/cast"
)

func setupGoCacheDriver[V any](t *testing.T) *GoCacheDriver[V] {
	memCache := gocache.New(5*time.Minute, 10*time.Minute)
	driverIndex = driverIndex + 1
	driverName := "mem_test" + cast.ToString(driverIndex)
	rand.Seed(time.Now().UnixNano())
	prefix := "cache_prefix"
	if rand.Intn(2) > 0 {
		prefix = ""
	}
	err := RegisterGoCacheDriver(driverName, memCache, prefix)
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
	return driver.(*GoCacheDriver[V])
}

func TestGoCacheDriver(t *testing.T) {
	goCacheDriver := setupGoCacheDriver[string](t)
	testCache[string](t, goCacheDriver, "test_string_key", "test_string_value")
	testNumberCache[string](t, goCacheDriver, "test_string_key", "test_string_value")

	goCacheDriverInt := setupGoCacheDriver[int](t)
	testNumberCache[int](t, goCacheDriverInt, "test_int_key", 2)

	goCacheDriverUint := setupGoCacheDriver[uint](t)
	testNumberCache[uint](t, goCacheDriverUint, "test_uint_key", uint(2))

	goCacheDriverFloat := setupGoCacheDriver[float32](t)
	testNumberCache[float32](t, goCacheDriverFloat, "test_float_key", float32(2.0))
}
