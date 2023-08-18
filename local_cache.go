package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type BuildInMapCacheOption func(cache *BuildInMapCache)

type BuildInMapCache struct {
	data        map[string]*item
	outInterval time.Duration
	mutex       *sync.RWMutex
	close       chan struct{}
	onEvicted   func(key string, val any)
}

type item struct {
	val      any
	deadline time.Time
}

func (i *item) deadlineBefore(time time.Time) bool {
	return i.deadline.IsZero() && i.deadline.Before(time)
}

func NewBuildInMapCache(cap int, opts ...BuildInMapCacheOption) *BuildInMapCache {
	cache := &BuildInMapCache{
		data:        make(map[string]*item, cap),
		outInterval: time.Hour,
		mutex:       &sync.RWMutex{},
		close:       make(chan struct{}),
		onEvicted:   func(key string, val any) {},
	}
	for _, opt := range opts {
		opt(cache)
	}

	go func() {
		ticker := time.NewTicker(cache.outInterval)
		for {
			select {
			case <-ticker.C:
				cache.mutex.Lock()
				cnt := 0
				for key, res := range cache.data {
					if cnt > 1000 {
						break
					}
					if res.deadlineBefore(time.Now()) {
						cache.delete(key)
					}
					cnt++
				}
				cache.mutex.Unlock()

			case <-cache.close:
				return
			}
		}
	}()
	return cache
}

func BuildInMapCacheWithOutInterval(interval time.Duration) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.outInterval = interval
	}
}

func BuildInMapCacheWithEvictedCallback(fn func(key string, val any)) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.onEvicted = fn
	}
}

func (c *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	c.data[key] = &item{
		val:      val,
		deadline: dl,
	}
	return nil
}

func (c *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	c.mutex.RLock()
	res, ok := c.data[key]
	c.mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
	}
	if res.deadlineBefore(time.Now()) {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		res, ok = c.data[key]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
		if res.deadlineBefore(time.Now()) {
			c.delete(key)
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
	}
	return res.val, nil
}

func (c *BuildInMapCache) LoadAndDelete(ctx context.Context, key string) (any, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	res, ok := c.data[key]
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
	}
	c.delete(key)
	return res.val, nil
}

func (c *BuildInMapCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.delete(key)
	return nil
}

func (c *BuildInMapCache) delete(key string) {
	val, ok := c.data[key]
	if !ok {
		return
	}
	delete(c.data, key)
	c.onEvicted(key, val)
}
