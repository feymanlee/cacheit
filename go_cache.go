package cacheit

import (
	"context"
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// GoCacheDriver go-cache driver implemented
type GoCacheDriver[V any] struct {
	baseDriver
}

func (d *GoCacheDriver[V]) Set(key string, value V, t time.Duration) error {
	d.memCache.Set(d.getCacheKey(key), value, t)
	return nil
}

func (d *GoCacheDriver[V]) SetMany(many []Many[V]) error {
	for _, item := range many {
		_ = d.Set(item.Key, item.Value, item.TTL)
	}
	return nil
}

func (d *GoCacheDriver[V]) Many(keys []string) (map[string]V, error) {
	items := make(map[string]V)
	for _, key := range keys {
		if value, found := d.memCache.Get(d.getCacheKey(key)); found {
			items[key] = value.(V)
		}
	}
	return items, nil
}

func (d *GoCacheDriver[V]) DelMany(keys []string) error {
	for _, key := range keys {
		d.memCache.Delete(d.getCacheKey(key))
	}
	return nil
}

func (d *GoCacheDriver[V]) ForgetMany(keys []string) error {
	return d.DelMany(keys)
}

func (d *GoCacheDriver[V]) Add(key string, value V, t time.Duration) error {
	err := d.memCache.Add(d.getCacheKey(key), value, t)
	if err != nil {
		return ErrCacheExisted
	}
	return err
}

func (d *GoCacheDriver[V]) Forever(key string, value V) error {
	d.memCache.Set(d.getCacheKey(key), value, gocache.NoExpiration)
	return nil
}

func (d *GoCacheDriver[V]) Forget(key string) error {
	d.memCache.Delete(d.getCacheKey(key))
	return nil
}

func (d *GoCacheDriver[V]) Del(key string) error {
	return d.Forget(key)
}

func (d *GoCacheDriver[V]) Flush() error {
	d.memCache.Flush()
	return nil
}

func (d *GoCacheDriver[V]) Get(key string) (result V, err error) {
	if value, found := d.memCache.Get(d.getCacheKey(key)); !found {
		return result, ErrCacheMiss
	} else {
		var ok bool
		result, ok = value.(V)
		if !ok {
			return result, fmt.Errorf("cache item type mismatch: expected %T, got %T", result, value)
		}
		return result, nil
	}
}

func (d *GoCacheDriver[V]) Has(key string) (bool, error) {
	_, found := d.memCache.Get(d.getCacheKey(key))
	return found, nil
}

func (d *GoCacheDriver[V]) SetNumber(key string, value V, t time.Duration) error {
	switch any(value).(type) {
	case int, int8, int16, int32, int64:
		d.memCache.Set(d.getCacheKey(key), cast.ToInt64(value), t)
	case uint, uint8, uint16, uint32, uint64:
		d.memCache.Set(d.getCacheKey(key), cast.ToUint64(value), t)
	case float32, float64:
		d.memCache.Set(d.getCacheKey(key), cast.ToFloat64(value), t)
	default:
		return fmt.Errorf("the value for %v is not a number", value)
	}
	return nil
}

func (d *GoCacheDriver[V]) Increment(key string, n V) (ret V, err error) {
	var res any
	switch any(n).(type) {
	case int, int8, int16, int32, int64:
		res, err = d.memCache.IncrementInt64(d.getCacheKey(key), cast.ToInt64(n))
	case uint, uint8, uint16, uint32, uint64:
		res, err = d.memCache.IncrementUint64(d.getCacheKey(key), cast.ToUint64(n))
	case float32, float64:
		res, err = d.memCache.IncrementFloat64(d.getCacheKey(key), cast.ToFloat64(n))
	default:
		return ret, fmt.Errorf("invalid number type: %T", n)
	}
	if err != nil {
		return
	}
	return toAnyE[V](res)
}

func (d *GoCacheDriver[V]) Decrement(key string, n V) (ret V, err error) {
	var res any
	switch any(n).(type) {
	case int, int8, int16, int32, int64:
		res, err = d.memCache.DecrementInt64(d.getCacheKey(key), cast.ToInt64(n))
	case uint, uint8, uint16, uint32, uint64:
		res, err = d.memCache.DecrementUint64(d.getCacheKey(key), cast.ToUint64(n))
	case float32, float64:
		res, err = d.memCache.DecrementFloat64(d.getCacheKey(key), cast.ToFloat64(n))
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

func (d *GoCacheDriver[V]) Remember(key string, ttl time.Duration, callback func() (V, error), force bool) (result V, err error) {
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

func (d *GoCacheDriver[V]) RememberForever(key string, callback func() (V, error), force bool) (V, error) {
	return d.Remember(key, gocache.NoExpiration, callback, force)
}

func (d *GoCacheDriver[V]) RememberMany(keys []string, ttl time.Duration, callback func(notHitKeys []string) (map[string]V, error), force bool) (map[string]V, error) {
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

func (d *GoCacheDriver[V]) TTL(key string) (ttl time.Duration, err error) {
	items := d.memCache.Items()
	if item, found := items[d.getCacheKey(key)]; found {
		if item.Expiration == 0 {
			return NoExpirationTTL, nil
		}
		now := time.Now()
		expirationTime := time.Unix(0, item.Expiration)
		if expirationTime.Before(now) {
			return ItemNotExistedTTL, fmt.Errorf("cached item %v expired", key)
		}
		return expirationTime.Sub(now), nil
	}
	return ItemNotExistedTTL, fmt.Errorf("cached item %v not found", key)
}

func (d *GoCacheDriver[V]) WithCtx(ctx context.Context) Driver[V] {
	d.ctx = ctx
	return d
}

func (d *GoCacheDriver[V]) WithSerializer(serializer Serializer) Driver[V] {
	d.serializer = serializer
	return d
}
