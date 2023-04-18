/**
 * @Author: lifameng@changba.com
 * @Description:
 * @File:  go_cache_test.go
 * @Date: 2023/4/11 22:49
 */
package cache

import (
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

func TestGoCacheDriver(t *testing.T) {
	memCache := gocache.New(5*time.Minute, 10*time.Minute)
	// 初始化 GoCacheDriver
	goCacheDriver, _ := New[string](DriverMemory, WithMemCache(memCache))
	testDriverString(t, goCacheDriver)
}
