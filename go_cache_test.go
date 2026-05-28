package cacheit

import (
	"context"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGoCacheDriver[V any](t *testing.T) *GoCacheDriver[V] {
	t.Helper()
	return setupGoCacheDriverWithPrefix[V](t, "cache_prefix")
}

func setupGoCacheDriverWithPrefix[V any](t *testing.T, prefix string) *GoCacheDriver[V] {
	t.Helper()

	memCache := gocache.New(5*time.Minute, 10*time.Minute)
	driverName := nextDriverName("mem_test")
	err := RegisterGoCacheDriver(driverName, memCache, prefix)
	require.NoError(t, err, "register go-cache driver")

	driver, err := Use[V](driverName)
	require.NoError(t, err, "use go-cache driver")

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

func TestGoCacheManyReturnsErrorOnTypeMismatch(t *testing.T) {
	memCache := gocache.New(5*time.Minute, 10*time.Minute)
	driverName := nextDriverName("mem_type_mismatch")

	require.NoError(t, RegisterGoCacheDriver(driverName, memCache, ""))

	stringDriver, err := Use[string](driverName)
	require.NoError(t, err)
	intDriver, err := Use[int](driverName)
	require.NoError(t, err)

	assert.NoError(t, stringDriver.Set("shared", "value", time.Minute))

	assert.NotPanics(t, func() {
		got, err := intDriver.Many([]string{"shared"})
		assert.Error(t, err)
		assert.Empty(t, got)
	})
}

func TestGoCacheFlushWithPrefixOnlyDeletesMatchingKeys(t *testing.T) {
	memCache := gocache.New(5*time.Minute, 10*time.Minute)
	driverNameA := nextDriverName("mem_prefix_a")
	driverNameB := nextDriverName("mem_prefix_b")

	require.NoError(t, RegisterGoCacheDriver(driverNameA, memCache, "prefix_a"))
	require.NoError(t, RegisterGoCacheDriver(driverNameB, memCache, "prefix_b"))

	driverA, err := Use[string](driverNameA)
	require.NoError(t, err)
	driverB, err := Use[string](driverNameB)
	require.NoError(t, err)

	assert.NoError(t, driverA.Set("shared", "a", time.Minute))
	assert.NoError(t, driverB.Set("shared", "b", time.Minute))
	memCache.Set("unprefixed", "raw", time.Minute)

	assert.NoError(t, driverA.Flush())

	_, err = driverA.Get("shared")
	assert.ErrorIs(t, err, ErrCacheMiss)

	gotB, err := driverB.Get("shared")
	assert.NoError(t, err)
	assert.Equal(t, "b", gotB)

	raw, found := memCache.Get("unprefixed")
	assert.True(t, found)
	assert.Equal(t, "raw", raw)
}

func TestGoCacheFlushWithoutPrefixDeletesAllKeys(t *testing.T) {
	driver := setupGoCacheDriverWithPrefix[string](t, "")

	assert.NoError(t, driver.Set("managed", "value", time.Minute))
	driver.memCache.Set("raw", "value", time.Minute)

	assert.NoError(t, driver.Flush())

	_, err := driver.Get("managed")
	assert.ErrorIs(t, err, ErrCacheMiss)

	_, found := driver.memCache.Get("raw")
	assert.False(t, found)
}

func TestGoCacheComplexNumberOperationsReturnError(t *testing.T) {
	driver := setupGoCacheDriver[complex64](t)
	key := "complex_number"
	value := complex64(1 + 2i)

	assert.Error(t, driver.SetNumber(key, value, time.Minute))
	_, err := driver.Increment(key, value)
	assert.Error(t, err)
	_, err = driver.Decrement(key, value)
	assert.Error(t, err)
}
