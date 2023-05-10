package cacheit

import (
	"context"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

func TestGoCacheDriver(t *testing.T) {
	memCache := gocache.New(5*time.Minute, 10*time.Minute)
	// 初始化 GoCacheDriver
	goCacheDriver, _ := New[string](DriverMemory, WithMemCache(memCache))
	goCacheDriver = goCacheDriver.WithCtx(context.Background())
	testCache[string](t, goCacheDriver, "test_string_key", "test_string_value")
	testNumberCache[string](t, goCacheDriver, "test_string_key", "test_string_value")

	goCacheDriverInt, _ := New[int](DriverMemory, WithMemCache(memCache))
	testNumberCache[int](t, goCacheDriverInt, "test_int_key", 2)

	goCacheDriverUint, _ := New[uint](DriverMemory, WithMemCache(memCache))
	testNumberCache[uint](t, goCacheDriverUint, "test_uint_key", uint(2))

	goCacheDriverFloat, _ := New[float32](DriverMemory, WithMemCache(memCache))
	testNumberCache[float32](t, goCacheDriverFloat, "test_float_key", float32(2.0))
}
