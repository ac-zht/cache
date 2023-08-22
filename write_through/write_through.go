package write_through

import (
	"context"
	"github.com/zht-account/cache"
	"log"
	"time"
)

type WriteThroughCache struct {
    cache.Cache
    storeFunc func(ctx context.Context, key string, val []byte) error
}

func NewWriteThroughCache(cache cache.Cache, storeFunc func(ctx context.Context, key string, val []byte) error) *WriteThroughCache {
    res := &WriteThroughCache{
        Cache:     cache,
        storeFunc: storeFunc,
    }
    return res
}

//Set 先写库再写缓存同步操作
func (c *WriteThroughCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
    if err := c.storeFunc(ctx, key, val); err != nil {
        return err
    }
    return c.Cache.Set(ctx, key, val, expiration)
}

//SemiAsyncSet 先写库再写缓存半异步操作
func (c *WriteThroughCache) SemiAsyncSet(ctx context.Context, key string, val []byte, expiration time.Duration) error {
    err := c.storeFunc(ctx, key, val)
    go func() {
        e := c.Cache.Set(ctx, key, val, expiration)
        if e != nil {
            log.Fatalln(e)
        }
    }()
    return err
}

//AsyncSet 先写库再写缓存全异步操作
func (c *WriteThroughCache) AsyncSet(ctx context.Context, key string, val []byte, expiration time.Duration) error {
    go func() {
        if err := c.storeFunc(ctx, key, val); err != nil {
            log.Fatalln(err)
        }
        if e := c.Cache.Set(ctx, key, val, expiration); e != nil {
            log.Fatalln(e)
        }
    }()
    return nil
}

//SetV2 先写缓存再写库同步操作
func (c *WriteThroughCache) SetV2(ctx context.Context, key string, val []byte, expiration time.Duration) error {
    if err := c.Cache.Set(ctx, key, val, expiration); err != nil {
        return err
    }
    return c.storeFunc(ctx, key, val)
}
