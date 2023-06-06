# cacheit

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/feymanlee/cacheit?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/feymanlee/cacheit)](https://goreportcard.com/report/github.com/feymanlee/cacheit)
[![Unit-Tests](https://github.com/feymanlee/logit/workflows/Unit-Tests/badge.svg)](https://github.com/feymanlee/cacheit/actions)
[![Coverage Status](https://coveralls.io/repos/github/feymanlee/cacheit/badge.svg?branch=main)](https://coveralls.io/github/feymanlee/cacheit?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/feymanlee/cacheit.svg)](https://pkg.go.dev/github.com/feymanlee/cacheit)
[![License](https://img.shields.io/github/license/feymanlee/cacheit)](./LICENSE)

GO 缓存库，支持 [go-redis](https://github.com/redis/go-redis/tree/v8)
和 [go-cache](https://github.com/patrickmn/go-cache) 作为缓存驱动。它提供了一个简单的接口，允许您轻松地在您的项目中实现缓存。

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
    // Items not found in the cache will have a zero value.
    Many(keys []string) (map[string]V, error)
    // SetNumber set the number value of an item in the cache.
    SetNumber(key string, value V, t time.Duration) error
    // Increment the value of an item in the cache.
    Increment(key string, n V) (V, error)
    // Decrement the value of an item in the cache.
    Decrement(key string, n V) (V, error)
    // Remember Get an item from the cache, or execute the given Closure and store the result.
    Remember(key string, ttl time.Duration, callback func () (V, error)) (V, error)
    // RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
    RememberForever(key string, callback func () (V, error)) (V, error)
    // TTL Get cache ttl
    TTL(key string) (time.Duration, error)
    // WithCtx with context
    WithCtx(ctx context.Context) Driver[V]
    // WithWithSerializer with cache serializer
    WithSerializer(serializer Serializer) Driver[V]
}
```

## Usage

### Base Usage

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/feymanlee/cacheit"
	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	err := cacheit.RegisterRedisDriver("redis_test", redisClient, "cache_prefix")
	if err != nil {
		log.Fatal(err)
	}
	// go-cache 客户端
	memCache := gocache.New(5*time.Minute, 10*time.Minute)
	err = cacheit.RegisterGoCacheDriver("memory_test", memCache, "cache_prefix")
	if err != nil {
		log.Fatal(err)
	}
	// set default cache
	cacheit.SetDefault("redis_test")

	driver, err := cacheit.Use[string]("memory_test")
	if err != nil {
		log.Fatal(err)
	}
	
	err = driver.Set("cache_key", "cache_value", time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	
	get, err := cacheit.UseDefault[string]().Get("cache_key")
	if err != nil {
		return
	}
	fmt.Println(get)
}
```

### Add
Store an item in the cache if the key doesn't exist.

```go
err = driver.Add("key", "value", time.Minute*10)
if err != nil {
log.Println("Error adding cache:", err)
}
```

### Set
Store an item in the cache for a given number of seconds.
```go
err = driver.Set("key2", "value2", time.Minute*5)
if err != nil {
log.Println("Error setting cache:", err)
}
```

### Get
Retrieve an item from the cache by key.

```go
value, err := driver.Get("key2")
if err != nil {
log.Println("Error getting cache:", err)
} else {
log.Println("Cache value:", value)
}
```

### SetMany
Store multiple items in the cache for a given number of seconds.
```go
err = driver.SetMany([]cacheit.Many[string]{
{Key: "key3", Value: "value3", TTL: time.Minute * 10},
{Key: "key4", Value: "value4", TTL: time.Minute * 15},
})
if err != nil {
log.Println("Error setting many caches:", err)
}
```

### Many
Retrieve multiple items from the cache by key. Items not found in the cache will have a zero value.

```go
values, err := driver.Many([]string{"key2", "key3", "key4"})
if err != nil {
log.Println("Error getting many caches:", err)
} else {
log.Println("Cache values:", values)
}
```

### Forever
Store an item in the cache indefinitely.

```go
err = driver.Forever("key5", "value5")
if err != nil {
log.Println("Error storing cache forever:", err)
}
```

### Forget
Remove an item from the cache.

```go
err = driver.Forget("key")
if err != nil {
log.Println("Error removing cache:", err)
}
```

### Flush
Remove all items from the cache.

```go
err = driver.Flush()
if err != nil {
log.Println("Error flushing cache:", err)
}
```

### Has
Determined if an item exists in the cache.

```go
has, err := driver.Has("key2")
if err != nil {
log.Println("Error checking if cache exists:", err)
} else if has {
log.Println("Cache key2 exists")
} else {
log.Println("Cache key2 does not exist")
}
```

### SetNumber
Set the number value of an item in the cache.

```go
err = driver.SetNumber("number_key", 1, time.Minute*10)
if err != nil {
log.Println("Error setting number cache:", err)
}
```

### Increment
Increment the value of an item in the cache.

```go
newValue, err := driver.Increment("number_key", 1)
if err != nil {
log.Println("Error incrementing cache value:", err)
} else {
log.Println("Incremented cache value:", newValue)
}
```

### Increment

```go
newValue, err := driver.Increment("number_key", 1)
if err != nil {
log.Println("Error incrementing cache value:", err)
} else {
log.Println("Incremented cache value:", newValue)
}
```

### Remember

```go
rememberValue, err := driver.Remember("remember_key", time.Minute*10, func () (string, error) {
time.Sleep(time.Millisecond * 50)
return "remember_value", nil
})
if err != nil {
log.Println("Error remembering cache:", err)
} else {
log.Println("Remember cache value:", rememberValue)
}
```

### RememberForever

```go
rememberForeverValue, err := driver.RememberForever("remember_forever_key", func() (string, error) {
time.Sleep(time.Millisecond * 50)
return "remember_forever_value", nil
})
if err != nil {
log.Println("Error remembering cache forever:", err)
} else {
log.Println("Remember cache value forever:", rememberForeverValue)
}
```

### TTL

```go
ttl, err := driver.TTL("key2")
if err != nil {
log.Println("Error getting cache TTL:", err)
} else {
log.Println("Cache TTL:", ttl)
}
```

### WithCtx

```go
err = driver.WithCtx(context.TODO()).Set("key2", "value2", time.Minute*5)
if err != nil {
log.Println("Error setting cache:", err)
}
```

## 更多功能

请查阅源代码以了解更多功能和用法。

## 贡献

欢迎向项目贡献代码、提交 bug 报告或提出新功能建议。请务必遵循贡献指南。

## 鸣谢

> [GoLand](https://www.jetbrains.com/go/?from=cacheit) A Go IDE with extended support for JavaScript, TypeScript, and databases。

特别感谢 [JetBrains](https://www.jetbrains.com/?from=cacheit) 为开源项目提供免费的 [GoLand](https://www.jetbrains.com/go/?from=cacheit) 等 IDE 的授权  
[<img src=".github/jetbrains-variant-3.png" width="200"/>](https://www.jetbrains.com/?from=cacheit)

## 许可证

本项目基于 MIT 许可证 发布。
