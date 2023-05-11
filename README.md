# cacheit
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/feymanlee/cacheit?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/feymanlee/cacheit)](https://goreportcard.com/report/github.com/feymanlee/cacheit)
[![Unit-Tests](https://github.com/feymanlee/logit/workflows/Unit-Tests/badge.svg)](https://github.com/feymanlee/cacheit/actions)
[![Coverage Status](https://coveralls.io/repos/github/feymanlee/cacheit/badge.svg?branch=main)](https://coveralls.io/github/feymanlee/cacheit?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/feymanlee/cacheit.svg)](https://pkg.go.dev/github.com/feymanlee/cacheit)

GO 缓存库，支持 go-redis 和 go-cache 作为缓存驱动。它提供了一个简单的接口，允许您轻松地在您的项目中实现缓存。


## 安装

使用以下命令将 `cacheit` 添加到您的 Go 项目中：
```sh
go get github.com/feymanlee/cacheit
```

## 接口定义
```go
type Driver[V any] interface {
    // Add Store an item in the cache if the key doesn't exist.
    Add(key string, value V, t time.Duration) error
    // Set Store an item in the cache for a given number of seconds.
    Set(key string, value V, t time.Duration) error
    // SetMany Store multiple items in the cache for a given number of seconds.
    SetMany(many []Many[V]) error
    // Forever Store an item in the cache indefinitely.
    Forever(key string, value V) error
    // Forget Remove an item from the cache.
    Forget(key string) error
    // Flush Remove all items from the cache.
    Flush() error
    // Get Retrieve an item from the cache by key.
    Get(key string) (V, error)
    // Has Determined if an item exists in the cache.
    Has(key string) (bool, error)
    // Many Retrieve multiple items from the cache by key.
    // Items not found in the cache will have a nil value.
    Many(keys []string) (map[string]V, error)
    // SetNumber set the int64 value of an item in the cache.
    SetNumber(key string, value V, t time.Duration) error
    // Increment the value of an item in the cache.
    Increment(key string, n V) (V, error)
    // Decrement the value of an item in the cache.
    Decrement(key string, n V) (V, error)
    // Remember Get an item from the cache, or execute the given Closure and store the result.
    Remember(key string, ttl time.Duration, callback func() (V, error)) (V, error)
    // RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
    RememberForever(key string, callback func() (V, error)) (V, error)
    // TTL Get cache ttl
    TTL(key string) (time.Duration, error)
    // WithCtx with context
    WithCtx(ctx context.Context) Driver[V]
}
```

## Usage
```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/feymanlee/cacheit"
)

func main() {
    // go-cache 客户端
	// memCache := gocache.New(5*time.Minute, 10*time.Minute)
	// 初始化 GoCacheDriver
	// driver, _ := New[string](DriverMemory, WithMemCache(memCache))
	
	// 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 创建缓存驱动实例
	driver, err := cacheit.New[string](cacheit.DriverRedis, cacheit.WithRedisClient(redisClient), cacheit.WithPrefix("myapp"))
	if err != nil {
		log.Fatal(err)
	}

	// 示例 1: 添加缓存
	err = driver.Add("key", "value", time.Minute*10)
	if err != nil {
		log.Println("Error adding cache:", err)
	}

	// 示例 2: 设置缓存
	err = driver.Set("key2", "value2", time.Minute*5)
	if err != nil {
		log.Println("Error setting cache:", err)
	}

	// 示例 3: 设置多个缓存
	err = driver.SetMany([]cacheit.Many[string]{
		{Key: "key3", Value: "value3", TTL: time.Minute * 10},
		{Key: "key4", Value: "value4", TTL: time.Minute * 15},
	})
	if err != nil {
		log.Println("Error setting many caches:", err)
	}

	// 示例 4: 永久存储缓存
	err = driver.Forever("key5", "value5")
	if err != nil {
		log.Println("Error storing cache forever:", err)
	}

	// 示例 5: 移除缓存
	err = driver.Forget("key")
	if err != nil {
		log.Println("Error removing cache:", err)
	}

	// 示例 6: 清空缓存
	err = driver.Flush()
	if err != nil {
		log.Println("Error flushing cache:", err)
	}

	// 示例 7: 获取缓存
	value, err := driver.Get("key2")
	if err != nil {
		log.Println("Error getting cache:", err)
	} else {
		log.Println("Cache value:", value)
	}

	// 示例 8: 判断缓存是否存在
	has, err := driver.Has("key2")
	if err != nil {
		log.Println("Error checking if cache exists:", err)
	} else if has {
		log.Println("Cache key2 exists")
	} else {
		log.Println("Cache key2 does not exist")
	}

	// 示例 9: 获取多个缓存
	values, err := driver.Many([]string{"key2", "key3", "key4"})
	if err != nil {
		log.Println("Error getting many caches:", err)
	} else {
		log.Println("Cache values:", values)
	}

	// 示例 10: 设置数字缓存
	err = driver.SetNumber("number_key", 1, time.Minute*10)
	if err != nil {
		log.Println("Error setting number cache:", err)
	}

	// 示例 11: 自增缓存值
	newValue, err := driver.Increment("number_key", 1)
	if err != nil {
		log.Println("Error incrementing cache value:", err)
	} else {
		log.Println("Incremented cache value:", newValue)
	}
	// 示例 12: 自减缓存值
	newValue, err = driver.Decrement("number_key", 1)
	if err != nil {
		log.Println("Error decrementing cache value:", err)
	} else {
		log.Println("Decremented cache value:", newValue)
	}

	// 示例 13: 获取或设置缓存
	rememberValue, err := driver.Remember("remember_key", time.Minute*10, func() (string, error) {
		// 模拟数据获取
		time.Sleep(time.Millisecond * 50)
		return "remember_value", nil
	})
	if err != nil {
		log.Println("Error remembering cache:", err)
	} else {
		log.Println("Remember cache value:", rememberValue)
	}

	// 示例 14: 获取或永久设置缓存
	rememberForeverValue, err := driver.RememberForever("remember_forever_key", func() (string, error) {
		// 模拟数据获取
		time.Sleep(time.Millisecond * 50)
		return "remember_forever_value", nil
	})
	if err != nil {
		log.Println("Error remembering cache forever:", err)
	} else {
		log.Println("Remember cache value forever:", rememberForeverValue)
	}

	// 示例 15: 获取缓存 TTL
	ttl, err := driver.TTL("key2")
	if err != nil {
		log.Println("Error getting cache TTL:", err)
	} else {
		log.Println("Cache TTL:", ttl)
	}
}
```

## 更多功能
请查阅源代码以了解更多关于 cacheit 的功能和用法。

## 贡献
欢迎向项目贡献代码、提交 bug 报告或提出新功能建议。请务必遵循贡献指南。

## 许可证
本项目基于 MIT 许可证 发布。
