package cacheit

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// RedisDriver go-redis driver implemented
type RedisDriver[V any] struct {
	baseDriver
}

func (d *RedisDriver[V]) Set(key string, value V, t time.Duration) error {
	serialize, err := d.serializer.Serialize(value)
	if err != nil {
		return err
	}

	return d.redisClient.Set(d.ctx, d.getCacheKey(key), string(serialize), t).Err()
}

func (d *RedisDriver[V]) SetMany(many []Many[V]) error {
	pipeline := d.redisClient.Pipeline()
	defer pipeline.Close()
	for _, m := range many {
		serialize, err := d.serializer.Serialize(m.Value)
		if err != nil {
			return err
		}
		pipeline.SetEX(d.ctx, d.getCacheKey(m.Key), serialize, m.TTL)
	}
	_, err := pipeline.Exec(d.ctx)
	return err
}

func (d *RedisDriver[V]) Many(keys []string) (map[string]V, error) {
	results := make(map[string]V)
	cacheKeys := lo.Map(keys, func(key string, index int) string {
		return d.getCacheKey(key)
	})
	result, err := d.redisClient.MGet(d.ctx, cacheKeys...).Result()
	if err != nil {
		return nil, err
	}
	for i, r := range result {
		if r == nil {
			continue
		}
		var v V
		err = d.serializer.UnSerialize([]byte(cast.ToString(r)), &v)
		if err != nil {
			continue
		}

		results[keys[i]] = v
	}

	return results, nil
}

func (d *RedisDriver[V]) Add(key string, value V, t time.Duration) error {
	serialize, err := d.serializer.Serialize(value)
	if err != nil {
		return err
	}
	res, err := d.redisClient.SetNX(d.ctx, d.getCacheKey(key), string(serialize), t).Result()
	if err != nil {
		return err
	}
	if !res {
		return ErrCacheExisted
	}
	return nil
}

func (d *RedisDriver[V]) Forever(key string, value V) error {
	serialize, err := d.serializer.Serialize(value)
	if err != nil {
		return err
	}
	return d.redisClient.Set(d.ctx, d.getCacheKey(key), string(serialize), -1).Err()
}

func (d *RedisDriver[V]) Forget(key string) error {
	return d.redisClient.Del(d.ctx, d.getCacheKey(key)).Err()
}

func (d *RedisDriver[V]) Flush() error {
	return d.redisClient.FlushDB(d.ctx).Err()
}

func (d *RedisDriver[V]) Get(key string) (V, error) {
	var result V
	if value, err := d.redisClient.Get(d.ctx, d.getCacheKey(key)).Result(); err != nil {
		if err == redis.Nil {
			return result, ErrCacheMiss
		}
		return result, err
	} else {
		err = d.serializer.UnSerialize([]byte(value), &result)
		return result, err
	}
}

func (d *RedisDriver[V]) Has(key string) (bool, error) {
	result, err := d.redisClient.Exists(d.ctx, d.getCacheKey(key)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, err
}

func (d *RedisDriver[V]) SetNumber(key string, value V, t time.Duration) error {
	if !isNumeric(value) {
		return fmt.Errorf("the value for %v is not a number", value)
	}
	return d.redisClient.Set(d.ctx, d.getCacheKey(key), value, t).Err()
}

func (d *RedisDriver[V]) Increment(key string, n V) (ret V, err error) {
	var res any
	switch reflect.TypeOf(n).Name() {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		res, err = d.redisClient.IncrBy(d.ctx, d.getCacheKey(key), cast.ToInt64(n)).Result()
	case "float32", "float64":
		res, err = d.redisClient.IncrByFloat(d.ctx, d.getCacheKey(key), cast.ToFloat64(n)).Result()
	default:
		var res V
		return res, fmt.Errorf("the value for %v is not a number", n)
	}
	if err != nil {
		return
	}
	ret, err = toAnyE[V](res)
	return
}

func (d *RedisDriver[V]) Decrement(key string, n V) (ret V, err error) {
	var res any
	switch reflect.TypeOf(n).Name() {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		res, err = d.redisClient.DecrBy(d.ctx, d.getCacheKey(key), cast.ToInt64(n)).Result()
	case "float32", "float64":
		res, err = d.redisClient.IncrByFloat(d.ctx, d.getCacheKey(key), 0-cast.ToFloat64(n)).Result()
	default:
		var res V
		return res, fmt.Errorf("the value for %v is not a number", n)
	}
	if err != nil {
		return
	}
	ret, err = toAnyE[V](res)
	return
}

func (d *RedisDriver[V]) Remember(key string, ttl time.Duration, callback func() (V, error)) (result V, err error) {
	if result, err = d.Get(key); err == nil {
		return
	} else {
		if result, err = callback(); err != nil {
			return
		}
		err = d.Set(key, result, ttl)
		return
	}
}

func (d *RedisDriver[V]) RememberForever(key string, callback func() (V, error)) (V, error) {
	return d.Remember(key, redis.KeepTTL, callback)
}

func (d *RedisDriver[V]) TTL(key string) (ttl time.Duration, err error) {
	return d.redisClient.TTL(d.ctx, d.getCacheKey(key)).Result()
}

func (d *RedisDriver[V]) WithCtx(ctx context.Context) Driver[V] {
	d.ctx = ctx
	return d
}
