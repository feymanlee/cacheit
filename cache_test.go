package cacheit

import (
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("unsupported driver", func(t *testing.T) {
		_, err := New[string](DriverType("unsupported"))
		assert.Error(t, err)
	})

	t.Run("memory driver", func(t *testing.T) {
		memCache := gocache.New(5*time.Minute, 10*time.Minute)
		driver, err := New[string](DriverMemory, WithMemCache(memCache))
		assert.NoError(t, err)
		assert.IsType(t, &GoCacheDriver[string]{}, driver)
	})
	t.Run("redis driver", func(t *testing.T) {
		driver, err := setupRedisDriver[string]()
		assert.NoError(t, err)
		assert.IsType(t, &RedisDriver[string]{}, driver)
	})
}

func testDriverString(t *testing.T, driver Driver[string]) {
	key := "test_key"
	value := "test_value"
	duration := time.Second * 10
	var err error
	t.Run("add", func(t *testing.T) {
		err = driver.Add(key, value, duration)
		assert.NoError(t, err)

		result, err := driver.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})
	t.Run("put and get", func(t *testing.T) {
		err = driver.Put(key, value, duration)
		assert.NoError(t, err)

		result, err := driver.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("forget", func(t *testing.T) {
		err = driver.Forget(key)
		assert.NoError(t, err)

		_, err = driver.Get(key)
		assert.Error(t, err)
	})
	// Add other test cases for the memory driver here.
}
