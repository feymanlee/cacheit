package cacheit

import (
	"context"
	"errors"
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
	if len(many) == 0 {
		return nil
	}
	pipeline := d.redisClient.Pipeline()
	defer pipeline.Close()
	for _, m := range many {
		serialize, err := d.serializer.Serialize(m.Value)
		if err != nil {
			return err
		}
		pipeline.Set(d.ctx, d.getCacheKey(m.Key), string(serialize), normalizeTTL(m.TTL))
	}
	_, err := pipeline.Exec(d.ctx)
	return err
}

func (d *RedisDriver[V]) Many(keys []string) (map[string]V, error) {
	results := make(map[string]V)
	if len(keys) == 0 {
		return results, nil
	}
	cacheKeys := d.getCacheKeys(keys)
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

func (d *RedisDriver[V]) DelMany(keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	cacheKeys := d.getCacheKeys(keys)
	return d.redisClient.Del(d.ctx, cacheKeys...).Err()
}

func (d *RedisDriver[V]) ForgetMany(keys []string) error {
	return d.DelMany(keys)
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
	return d.redisClient.Set(d.ctx, d.getCacheKey(key), string(serialize), 0).Err()
}

func (d *RedisDriver[V]) Forget(key string) error {
	return d.redisClient.Del(d.ctx, d.getCacheKey(key)).Err()
}

func (d *RedisDriver[V]) Del(key string) error {
	return d.Forget(key)
}

func (d *RedisDriver[V]) Flush() error {
	if d.prefix != "" {
		var cursor uint64
		pattern := d.prefix + ":*"
		for {
			keys, nextCursor, err := d.redisClient.Scan(d.ctx, cursor, pattern, 0).Result()
			if err != nil {
				return err
			}
			if len(keys) > 0 {
				if err := d.redisClient.Del(d.ctx, keys...).Err(); err != nil {
					return err
				}
			}
			if nextCursor == 0 {
				return nil
			}
			cursor = nextCursor
		}
	}
	return d.redisClient.FlushDB(d.ctx).Err()
}

func (d *RedisDriver[V]) Get(key string) (V, error) {
	var result V
	if value, err := d.redisClient.Get(d.ctx, d.getCacheKey(key)).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
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

func (d *RedisDriver[V]) Remember(key string, ttl time.Duration, callback func() (V, error), force bool) (result V, err error) {
	if !force {
		if result, err = d.Get(key); err == nil {
			return
		}
	}
	if result, err = callback(); err != nil {
		return
	}
	err = d.Set(key, result, ttl)
	return
}

func (d *RedisDriver[V]) RememberForever(key string, callback func() (V, error), force bool) (V, error) {
	return d.Remember(key, 0, callback, force)
}

func (d *RedisDriver[V]) RememberMany(keys []string, ttl time.Duration, callback func(notHitKeys []string) (map[string]V, error), force bool) (map[string]V, error) {
	var (
		notHitKeys []string
		err        error
	)
	many := make(map[string]V)
	if !force {
		many, err = d.Many(keys)
		if err != nil {
			return nil, err
		}
		notHitKeys = lo.Without(keys, lo.Keys(many)...)
		if len(notHitKeys) == 0 {
			return many, nil
		}
	} else {
		notHitKeys = keys
	}
	notCacheItems, err := callback(notHitKeys)
	if err != nil {
		return nil, err
	}
	var needCacheItems []Many[V]
	for s, v := range notCacheItems {
		needCacheItems = append(needCacheItems, Many[V]{
			Key:   s,
			Value: v,
			TTL:   ttl,
		})
	}
	err = d.SetMany(needCacheItems)
	if err != nil {
		return nil, err
	}
	return lo.Assign(many, notCacheItems), nil
}

func (d *RedisDriver[V]) TTL(key string) (ttl time.Duration, err error) {
	return d.redisClient.TTL(d.ctx, d.getCacheKey(key)).Result()
}

func (d *RedisDriver[V]) WithCtx(ctx context.Context) Driver[V] {
	d.ctx = ctx
	return d
}

func (d *RedisDriver[V]) WithSerializer(serializer Serializer) Driver[V] {
	d.serializer = serializer
	return d
}

func normalizeTTL(ttl time.Duration) time.Duration {
	if ttl == NoExpirationTTL {
		return 0
	}
	return ttl
}
