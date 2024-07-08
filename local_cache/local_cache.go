package local_cache

import (
	"context"
	"fmt"
	"github.com/ac-zht/cache"
	"sync"
	"time"
)

type BuildInMapCacheOption func(cache *BuildInMapCache)

type BuildInMapCache struct {
	Data        map[string]*item
	outInterval time.Duration
	Mutex       *sync.RWMutex
	close       chan struct{}
	OnEvicted   func(key string, val any)
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
		Data:        make(map[string]*item, cap),
		outInterval: time.Hour,
		Mutex:       &sync.RWMutex{},
		close:       make(chan struct{}),
		OnEvicted:   func(key string, val any) {},
	}
	for _, opt := range opts {
		opt(cache)
	}

	go func() {
		ticker := time.NewTicker(cache.outInterval)
		for {
			select {
			case <-ticker.C:
				cache.Mutex.Lock()
				cnt := 0
				for key, res := range cache.Data {
					if cnt > 1000 {
						break
					}
					if res.deadlineBefore(time.Now()) {
						cache.delete(key)
					}
					cnt++
				}
				cache.Mutex.Unlock()

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
		cache.OnEvicted = fn
	}
}

func (c *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	c.Data[key] = &item{
		val:      val,
		deadline: dl,
	}
	return nil
}

func (c *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	c.Mutex.RLock()
	res, ok := c.Data[key]
	c.Mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, key)
	}
	if res.deadlineBefore(time.Now()) {
		c.Mutex.Lock()
		defer c.Mutex.Unlock()
		res, ok = c.Data[key]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, key)
		}
		if res.deadlineBefore(time.Now()) {
			c.delete(key)
			return nil, fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, key)
		}
	}
	return res.val, nil
}

func (c *BuildInMapCache) LoadAndDelete(ctx context.Context, key string) (any, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	res, ok := c.Data[key]
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, key)
	}
	c.delete(key)
	return res.val, nil
}

func (c *BuildInMapCache) Delete(ctx context.Context, key string) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.delete(key)
	return nil
}

func (c *BuildInMapCache) delete(key string) {
	val, ok := c.Data[key]
	if !ok {
		return
	}
	delete(c.Data, key)
	c.OnEvicted(key, val)
}
