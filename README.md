# cacheit

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/feymanlee/cacheit?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/feymanlee/cacheit)](https://goreportcard.com/report/github.com/feymanlee/cacheit)
[![Unit-Tests](https://github.com/feymanlee/cacheit/workflows/Unit-Tests/badge.svg)](https://github.com/feymanlee/cacheit/actions)
[![Coverage Status](https://coveralls.io/repos/github/feymanlee/cacheit/badge.svg?branch=main)](https://coveralls.io/github/feymanlee/cacheit?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/feymanlee/cacheit.svg)](https://pkg.go.dev/github.com/feymanlee/cacheit)
[![License](https://img.shields.io/github/license/feymanlee/cacheit)](./LICENSE)

`cacheit` 是一个轻量级 Go 缓存库，基于泛型提供统一的缓存接口，支持 [go-redis v8](https://github.com/redis/go-redis/tree/v8) 和 [go-cache](https://github.com/patrickmn/go-cache) 两种驱动。

## Features

- Redis 和本地内存缓存使用同一套 API。
- 支持泛型读写，减少业务代码里的类型转换。
- 支持 key prefix，便于多个业务模块共享同一个 Redis DB 或 go-cache 实例。
- 支持单个/批量读写、删除、TTL 查询、数值自增自减。
- 支持 `Remember` / `RememberForever` 缓存回源模式。
- Redis 驱动支持 `context.Context` 和自定义序列化器。

## Installation

```sh
go get github.com/feymanlee/cacheit
```

## Quick Start

### Redis

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/feymanlee/cacheit"
	"github.com/go-redis/redis/v8"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := cacheit.RegisterRedisDriver("redis", redisClient, "app_cache"); err != nil {
		log.Fatal(err)
	}

	driver, err := cacheit.Use[string]("redis")
	if err != nil {
		log.Fatal(err)
	}

	if err := driver.Set("hello", "world", time.Minute); err != nil {
		log.Fatal(err)
	}

	value, err := driver.Get("hello")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value)
}
```

### Memory

```go
package main

import (
	"log"
	"time"

	"github.com/feymanlee/cacheit"
	gocache "github.com/patrickmn/go-cache"
)

func main() {
	memCache := gocache.New(5*time.Minute, 10*time.Minute)

	if err := cacheit.RegisterGoCacheDriver("memory", memCache, "app_cache"); err != nil {
		log.Fatal(err)
	}

	driver, err := cacheit.Use[int]("memory")
	if err != nil {
		log.Fatal(err)
	}

	if err := driver.Set("count", 1, time.Minute); err != nil {
		log.Fatal(err)
	}
}
```

## Register Drivers

```go
err := cacheit.RegisterRedisDriver("redis", redisClient, "cache_prefix")
err = cacheit.RegisterGoCacheDriver("memory", memCache, "cache_prefix")
```

`driverName` 必须唯一，重复注册会返回错误。`cacheKeyPrefix` 会自动拼接到实际缓存 key 前面，例如业务 key `user:1` 会写成 `cache_prefix:user:1`。

也可以设置默认 driver：

```go
cacheit.SetDefault("redis")
driver := cacheit.UseDefault[string]()
```

如果默认 driver 未设置或不存在，`UseDefault` 会 panic。更推荐在业务代码中使用 `Use` 并显式处理错误。

## API

```go
type Driver[V any] interface {
	Add(key string, value V, t time.Duration) error
	Set(key string, value V, t time.Duration) error
	SetMany(many []Many[V]) error
	Forever(key string, value V) error
	Forget(key string) error
	Del(key string) error
	Flush() error
	Get(key string) (V, error)
	Has(key string) (bool, error)
	Many(keys []string) (map[string]V, error)
	DelMany(keys []string) error
	ForgetMany(keys []string) error
	SetNumber(key string, value V, t time.Duration) error
	Increment(key string, n V) (V, error)
	Decrement(key string, n V) (V, error)
	Remember(key string, ttl time.Duration, callback func() (V, error), force bool) (V, error)
	RememberForever(key string, callback func() (V, error), force bool) (V, error)
	RememberMany(keys []string, ttl time.Duration, callback func(notHitKeys []string) (map[string]V, error), force bool) (map[string]V, error)
	TTL(key string) (time.Duration, error)
	WithCtx(ctx context.Context) Driver[V]
	WithSerializer(serializer Serializer) Driver[V]
}
```

### TTL Semantics

- `Set(key, value, ttl)`：写入一个带 TTL 的缓存。
- `Set(key, value, 0)`：写入不过期缓存。
- `SetMany` 中 `TTL: 0` 和 `TTL: cacheit.NoExpirationTTL` 都表示不过期。
- `Forever` / `RememberForever`：写入不过期缓存。
- `TTL` 返回 `cacheit.NoExpirationTTL` 表示 key 存在且不过期。
- `TTL` 返回 `cacheit.ItemNotExistedTTL` 表示 key 不存在。

Redis 驱动使用 go-redis v8：`expiration == 0` 表示不过期，`redis.KeepTTL` 表示保留已有 TTL，二者语义不同。

### Prefix And Flush

注册 driver 时传入 prefix 后，`Flush` 只会清理当前 prefix 下的缓存：

```go
_ = cacheit.RegisterRedisDriver("user_cache", redisClient, "user")
_ = cacheit.RegisterRedisDriver("order_cache", redisClient, "order")

userCache, _ := cacheit.Use[string]("user_cache")
orderCache, _ := cacheit.Use[string]("order_cache")

_ = userCache.Set("1", "alice", time.Minute)
_ = orderCache.Set("1", "order-1", time.Minute)

_ = userCache.Flush() // only deletes user:*
```

如果 driver 没有 prefix，`Flush` 会清理整个 Redis DB 或整个 go-cache 实例。生产环境建议始终使用 prefix。

### Error Values

```go
var (
	ErrCacheMiss    = errors.New("cache not exists")
	ErrCacheExisted = errors.New("cache already existed")
)
```

- `Get` 在 key 不存在时返回 `ErrCacheMiss`。
- `Add` 在 key 已存在时返回 `ErrCacheExisted`。
- GoCache 驱动在泛型类型不匹配时返回错误，而不是 panic。

## Usage Examples

### Add

`Add` 只在 key 不存在时写入：

```go
err := driver.Add("key", "value", 10*time.Minute)
if errors.Is(err, cacheit.ErrCacheExisted) {
	log.Println("cache already exists")
}
```

### Set And Get

```go
if err := driver.Set("key", "value", 5*time.Minute); err != nil {
	log.Println("set cache:", err)
}

value, err := driver.Get("key")
if errors.Is(err, cacheit.ErrCacheMiss) {
	log.Println("cache miss")
} else if err != nil {
	log.Println("get cache:", err)
} else {
	log.Println("cache value:", value)
}
```

### SetMany And Many

```go
err := driver.SetMany([]cacheit.Many[string]{
	{Key: "key1", Value: "value1", TTL: time.Minute},
	{Key: "key2", Value: "value2", TTL: cacheit.NoExpirationTTL},
})
if err != nil {
	log.Println("set many:", err)
}

values, err := driver.Many([]string{"key1", "key2", "missing"})
if err != nil {
	log.Println("get many:", err)
}

// Missing keys are omitted from the returned map.
log.Println(values)
```

### Remember

`Remember` 会优先读缓存；未命中或 `force == true` 时执行 callback，并把 callback 返回值写入缓存。

```go
value, err := driver.Remember("profile:1", 10*time.Minute, func() (string, error) {
	time.Sleep(50 * time.Millisecond)
	return "profile payload", nil
}, false)
if err != nil {
	log.Println("remember:", err)
}

log.Println(value)
```

### RememberForever

```go
value, err := driver.RememberForever("config", func() (string, error) {
	return "config payload", nil
}, false)
if err != nil {
	log.Println("remember forever:", err)
}

log.Println(value)
```

### RememberMany

```go
keys := []string{"user:1", "user:2"}

users, err := driver.RememberMany(keys, time.Minute, func(notHitKeys []string) (map[string]string, error) {
	result := make(map[string]string, len(notHitKeys))
	for _, key := range notHitKeys {
		result[key] = "value for " + key
	}
	return result, nil
}, false)
if err != nil {
	log.Println("remember many:", err)
}

log.Println(users)
```

### Number Operations

Number operations support integer, unsigned integer, `float32`, and `float64` values. Complex numbers are not supported.

```go
counter, err := cacheit.Use[int]("redis")
if err != nil {
	log.Fatal(err)
}

if err := counter.SetNumber("counter", 1, time.Minute); err != nil {
	log.Println("set number:", err)
}

newValue, err := counter.Increment("counter", 2)
if err != nil {
	log.Println("increment:", err)
}

log.Println(newValue)
```

### TTL

```go
ttl, err := driver.TTL("key")
if err != nil {
	log.Println("ttl:", err)
}

switch ttl {
case cacheit.NoExpirationTTL:
	log.Println("key exists without expiration")
case cacheit.ItemNotExistedTTL:
	log.Println("key does not exist")
default:
	log.Println("ttl:", ttl)
}
```

### Context

`WithCtx` 对 Redis 驱动特别有用，可以为网络操作设置超时或取消信号：

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

err := driver.WithCtx(ctx).Set("key", "value", time.Minute)
if err != nil {
	log.Println("set cache:", err)
}
```

### Custom Serializer

Redis 驱动默认使用 JSON 序列化。可以实现 `Serializer` 接口替换序列化方式：

```go
type Serializer interface {
	Serialize(v any) ([]byte, error)
	UnSerialize(data []byte, v any) error
}

driver = driver.WithSerializer(&cacheit.JSONSerializer{})
```

## Development

当前仓库使用 Go modules：

```sh
go test ./...
go test -race ./...
go vet ./...
```

如果本地设置了不匹配的 `GOROOT`，可以先取消它：

```sh
env -u GOROOT go test ./...
```

## Acknowledgements

> [GoLand](https://www.jetbrains.com/go/?from=cacheit) A Go IDE with extended support for JavaScript, TypeScript, and databases.

特别感谢 [JetBrains](https://www.jetbrains.com/?from=cacheit) 为开源项目提供免费的 [GoLand](https://www.jetbrains.com/go/?from=cacheit) 等 IDE 授权。

[<img src="https://account.jetbrains.com/static/images/jetbrains-logo-inv.svg" width="150"/>](https://www.jetbrains.com/?from=cacheit)

## License

This project is released under the MIT License.
