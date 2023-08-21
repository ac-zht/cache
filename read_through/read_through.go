package read_through

import (
	"context"
	"fmt"
	"github.com/zht-account/cache"
	"log"
	"time"
)

type ReadThroughCache struct {
	cache.Cache
	expiration time.Duration
	loadFunc   func(ctx context.Context, key string) ([]byte, error)
	storeFunc  func(ctx context.Context, key string, val []byte) error
	deleteFunc func(ctx context.Context, key string) error
}

func NewReadThroughCache(cache cache.Cache, expiration time.Duration,
	loadFunc func(ctx context.Context, key string) ([]byte, error),
	storeFunc func(ctx context.Context, key string, val []byte) error,
	deleteFunc func(ctx context.Context, key string) error) *ReadThroughCache {
	res := &ReadThroughCache{
		Cache:      cache,
		expiration: expiration,
		loadFunc:   loadFunc,
		storeFunc:  storeFunc,
		deleteFunc: deleteFunc,
	}
	return res
}

func (c *ReadThroughCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.Cache.Get(ctx, key)
	if err == cache.ErrKeyNotFound {
		if val, err = c.loadFunc(ctx, key); err == nil {
			err = c.Cache.Set(ctx, key, val, c.expiration)
			if err != nil {
				return val, fmt.Errorf("read_through: refresh cache fail, key : %s", key)
			}
		}
	}
	return val, err
}

func (c *ReadThroughCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	err := c.storeFunc(ctx, key, val)
	if err != nil {
		return fmt.Errorf("read_through: database store fail, key : %s", key)
	}
	err = c.Cache.Set(ctx, key, val, expiration)
	if err != nil {
		return fmt.Errorf("read_through: cache store fail, key : %s", key)
	}
	return nil
}

func (c *ReadThroughCache) AsyncSet(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	err := c.storeFunc(ctx, key, val)
	if err != nil {
		return fmt.Errorf("read_through: database store fail, key : %s", key)
	}
	go func() {
		if err = c.Cache.Set(ctx, key, val, expiration); err != nil {
			log.Fatalf("read_through: cache sync store fail, key : %s", key)
		}
	}()
	return nil
}

func (c *ReadThroughCache) Delete(ctx context.Context, key string, val []byte) error {
	err := c.deleteFunc(ctx, key)
	if err != nil {
		return fmt.Errorf("read_through: database delete fail, key : %s", key)
	}
	err = c.Cache.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("read_through: cache delete fail, key : %s", key)
	}
	return nil
}

func (c *ReadThroughCache) AsyncDelete(ctx context.Context, key string) error {
	err := c.deleteFunc(ctx, key)
	if err != nil {
		return fmt.Errorf("read_through: database delete fail, key : %s", key)
	}
	go func() {
		if err = c.Cache.Delete(ctx, key); err != nil {
			log.Fatalf("read_through: cache async delete fail, key : %s", key)
		}
	}()
	return nil
}
