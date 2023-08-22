package read_through

import (
	"context"
	"github.com/zht-account/cache"
	"log"
	"time"
)

type ReadThroughCache struct {
	cache.Cache
	expiration time.Duration
	LoadFunc   func(ctx context.Context, key string) ([]byte, error)
}

func NewReadThroughCache(cache cache.Cache, expiration time.Duration,
	LoadFunc func(ctx context.Context, key string) ([]byte, error)) *ReadThroughCache {
	res := &ReadThroughCache{
		Cache:      cache,
		expiration: expiration,
		LoadFunc:   LoadFunc,
	}
	return res
}

// Get 同步操作
func (c *ReadThroughCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.Cache.Get(ctx, key)
	if err == cache.ErrKeyNotFound {
		if val, err = c.LoadFunc(ctx, key); err == nil {
			err2 := c.Cache.Set(ctx, key, val, c.expiration)
			if err2 != nil {
				return val, cache.NewErrRefreshCacheFail(key)
			}
		}
	}
	return val, err
}

// SemiAsyncGet 半异步操作
func (c *ReadThroughCache) SemiAsyncGet(ctx context.Context, key string) ([]byte, error) {
	val, err := c.Cache.Get(ctx, key)
	if err == cache.ErrKeyNotFound {
		if val, err = c.LoadFunc(ctx, key); err == nil {
			go func() {
				err2 := c.Cache.Set(ctx, key, val, c.expiration)
				if err2 != nil {
					log.Fatalln(err2)
				}
			}()
		}
	}
	return val, err
}

// AsyncGet 全异步操作
func (c *ReadThroughCache) AsyncGet(ctx context.Context, key string) ([]byte, error) {
	val, err := c.Cache.Get(ctx, key)
	if err == cache.ErrKeyNotFound {
		go func() {
			if data, e2 := c.LoadFunc(ctx, key); e2 == nil {
				e3 := c.Cache.Set(ctx, key, data, c.expiration)
				if e3 != nil {
					log.Fatalf("cache: refresh cache fail, %s", e3)
				}
			}
		}()
	}
	return val, err
}
