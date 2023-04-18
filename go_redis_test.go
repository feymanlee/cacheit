/**
 * @Author: lifameng@changba.com
 * @Description:
 * @File:  go_redis_test.go
 * @Date: 2023/4/11 22:48
 */

package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func setupRedisDriver[V any]() (*RedisDriver[V], error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	opt, err := redis.ParseURL("redis://" + mr.Addr())
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	driver, err := New[V](DriverRedis, WithRedisClient(client))
	if err != nil {
		return nil, err
	}

	return driver.(*RedisDriver[V]), nil
}

func TestRedisDriver(t *testing.T) {
	redisDriver, err := setupRedisDriver[string]()
	if err != nil {
		t.Fatalf("Failed to setup RedisDriver: %v", err)
	}
	testDriverString(t, redisDriver)
}
