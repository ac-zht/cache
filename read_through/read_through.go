package read_through

import "C"
import (
    "context"
    "github.com/zht-account/cache"
    "log"
    "time"
)

type ReadThroughCache struct {
    cache.Cache
    expiration time.Duration
    loadFunc   func(ctx context.Context, key string) ([]byte, error)
}

func NewReadThroughCache(cache cache.Cache, expiration time.Duration,
    loadFunc func(ctx context.Context, key string) ([]byte, error)) *ReadThroughCache {
    res := &ReadThroughCache{
        Cache:      cache,
        expiration: expiration,
        loadFunc:   loadFunc,
    }
    return res
}

func (c *ReadThroughCache) Get(ctx context.Context, key string) ([]byte, error) {
    val, err := c.Cache.Get(ctx, key)
    if err == cache.ErrKeyNotFound {
        if val, err = c.loadFunc(ctx, key); err == nil {
            err2 := c.Cache.Set(ctx, key, val, c.expiration)
            if err2 != nil {
                return val, cache.NewErrRefreshCacheFail(key)
            }
        }
    }
    return val, err
}

func (c *ReadThroughCache) AsyncGet(ctx context.Context, key string) ([]byte, error) {
    val, err := c.Cache.Get(ctx, key)
    if err == cache.ErrKeyNotFound {
        go func() {
            if data, e2 := c.loadFunc(ctx, key); e2 == nil {
                e3 := c.Cache.Set(ctx, key, data, c.expiration)
                if e3 != nil {
                    log.Fatalf("cache: refresh cache fail, %s", e3)
                }
            }
        }()
    }
    return val, err
}
