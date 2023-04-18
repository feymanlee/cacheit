/**
 * @Author: lifameng@changba.com
 * @Description:
 * @File:  memory
 * @Date: 2023/4/5 13:20
 */

package cacheit

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type GoCacheDriver[V any] struct {
	baseDriver
}

func (c *GoCacheDriver[V]) Put(key string, value V, t time.Duration) error {
	c.memCache.Set(key, value, t)
	return nil
}

func (c *GoCacheDriver[V]) PutMany(many []Many[V]) error {
	for _, item := range many {
		c.memCache.Set(item.Key, item.Value, item.TTL)
	}
	return nil
}

func (c *GoCacheDriver[V]) Many(keys []string) (map[string]V, error) {
	var zeroValue V
	items := make(map[string]V)
	for _, key := range keys {
		if value, found := c.memCache.Get(key); found {
			items[key] = value.(V)
		} else {
			items[key] = zeroValue
		}
	}
	return items, nil
}

func (g *GoCacheDriver[V]) Add(key string, value V, t time.Duration) error {
	err := g.memCache.Add(g.getCacheKey(key), value, t)
	return err
}

func (g *GoCacheDriver[V]) Forever(key string, value V) error {
	return g.memCache.Add(g.getCacheKey(key), value, gocache.NoExpiration)
}

func (g *GoCacheDriver[V]) Forget(key string) error {
	g.memCache.Delete(g.getCacheKey(key))
	return nil
}

func (g *GoCacheDriver[V]) Flush() error {
	g.memCache.Flush()
	return nil
}

func (g *GoCacheDriver[V]) Get(key string) (result V, err error) {
	if value, found := g.memCache.Get(g.getCacheKey(key)); !found {
		return result, fmt.Errorf("key not found")
	} else {
		var ok bool
		result, ok = value.(V)
		if !ok {
			return result, fmt.Errorf("type assertion failed")
		}
		return result, nil
	}
}

func (g *GoCacheDriver[V]) Has(key string) (bool, error) {
	_, found := g.memCache.Get(g.getCacheKey(key))
	return found, nil
}

func (g *GoCacheDriver[V]) SetInt64(key string, value int64, t time.Duration) error {
	err := g.memCache.Add(g.getCacheKey(key), value, t)
	return err
}

func (g *GoCacheDriver[V]) IncrementInt64(key string, value int64) (int64, error) {
	return g.memCache.IncrementInt64(g.getCacheKey(key), value)
}

func (g *GoCacheDriver[V]) DecrementInt64(key string, value int64) (int64, error) {
	return g.memCache.DecrementInt64(g.getCacheKey(key), value)
}

func (g *GoCacheDriver[V]) Remember(key string, ttl time.Duration, callback func() (V, error)) (result V, err error) {
	if result, err = g.Get(key); err == nil {
		return
	} else {
		if result, err = callback(); err != nil {
			err = g.Add(key, result, ttl)
		}
		return
	}
}

func (g *GoCacheDriver[V]) RememberForever(key string, callback func() (V, error)) (V, error) {
	return g.Remember(key, gocache.NoExpiration, callback)
}
