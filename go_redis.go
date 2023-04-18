/**
 * @Author: lifameng@changba.com
 * @Description:
 * @File:  redis
 * @Date: 2023/4/5 10:34
 */

package cache

import (
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisDriver[V any] struct {
	baseDriver
}

func (rc *RedisDriver[V]) Put(key string, value V, t time.Duration) error {
	serialize, err := rc.serializer.Serialize(value)
	if err != nil {
		return err
	}
	return rc.redisClient.SetEX(rc.ctx, rc.getCacheKey(key), string(serialize), t).Err()
}

func (rc *RedisDriver[V]) PutMany(many []Many[V]) error {
	pipeline := rc.redisClient.Pipeline()
	defer pipeline.Close()
	for _, m := range many {
		serialize, err := rc.serializer.Serialize(m.Value)
		if err != nil {
			return err
		}
		pipeline.SetEX(rc.ctx, rc.getCacheKey(m.Key), serialize, m.TTL)
	}
	_, err := pipeline.Exec(rc.ctx)
	return err
}

func (rc *RedisDriver[V]) Many(keys []string) (map[string]V, error) {
	results := make(map[string]V)
	pipe := rc.redisClient.Pipeline()
	defer pipe.Close()

	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(rc.ctx, key)
	}

	_, err := pipe.Exec(rc.ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for i, cmd := range cmds {
		value, err := cmd.Result()
		if err != nil && err != redis.Nil {
			continue
		}
		var v V
		err = rc.serializer.UnSerialize([]byte(value), &v)
		if err != nil {
			continue
		}

		results[keys[i]] = v
	}

	return results, nil
}

func (rc *RedisDriver[V]) Add(key string, value V, t time.Duration) error {
	serialize, err := rc.serializer.Serialize(value)
	if err != nil {
		return err
	}
	return rc.redisClient.SetNX(rc.ctx, rc.getCacheKey(key), string(serialize), t).Err()
}

func (rc *RedisDriver[V]) Forever(key string, value V) error {
	serialize, err := rc.serializer.Serialize(value)
	if err != nil {
		return err
	}
	return rc.redisClient.Set(rc.ctx, rc.getCacheKey(key), string(serialize), -1).Err()
}

func (rc *RedisDriver[V]) Forget(key string) error {
	return rc.redisClient.Del(rc.ctx, rc.getCacheKey(key)).Err()
}

func (rc *RedisDriver[V]) Flush() error {
	return nil
}

func (rc *RedisDriver[V]) Get(key string) (V, error) {
	var result V
	if value, err := rc.redisClient.Get(rc.ctx, rc.getCacheKey(key)).Result(); err != nil {
		return result, err
	} else {
		err = rc.serializer.UnSerialize([]byte(value), &result)
		return result, err
	}
}

func (rc *RedisDriver[V]) Has(key string) (bool, error) {
	result, err := rc.redisClient.Exists(rc.ctx, rc.getCacheKey(key)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, err
}

func (rc *RedisDriver[V]) SetInt64(key string, value int64, t time.Duration) error {
	return rc.redisClient.Set(rc.ctx, rc.getCacheKey(key), strconv.FormatInt(value, 10), t).Err()
}

func (rc *RedisDriver[V]) IncrementInt64(key string, value int64) (int64, error) {
	return rc.redisClient.IncrBy(rc.ctx, rc.getCacheKey(key), value).Result()
}

func (rc *RedisDriver[V]) DecrementInt64(key string, value int64) (int64, error) {
	return rc.redisClient.DecrBy(rc.ctx, key, value).Result()
}

func (rc *RedisDriver[V]) Remember(key string, ttl time.Duration, callback func() (V, error)) (result V, err error) {
	if result, err = rc.Get(key); err == nil {
		return
	} else {
		result, err = callback()
		if err != nil {
			return
		}
		err = rc.Add(key, result, ttl)
		return
	}
}

func (rc *RedisDriver[V]) RememberForever(key string, callback func() (V, error)) (V, error) {
	return rc.Remember(key, -1, callback)
}
