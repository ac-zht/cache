package read_through

import (
	"context"
	"fmt"
	"github.com/ac-zht/cache"
	"golang.org/x/sync/singleflight"
	"time"
)

type SingleflightCache struct {
	ReadThroughCache
	g *singleflight.Group
}

func NewSingleflightCache(cache cache.Cache, expiration time.Duration,
	LoadFunc func(ctx context.Context, key string) ([]byte, error)) *SingleflightCache {
	return &SingleflightCache{
		ReadThroughCache: ReadThroughCache{
			Cache:      cache,
			expiration: expiration,
			LoadFunc:   LoadFunc,
		},
		g: &singleflight.Group{},
	}
}

func (s *SingleflightCache) Get(ctx context.Context, key string) ([]byte, error) {
	v1, e1 := s.Cache.Get(ctx, key)
	if e1 == cache.ErrKeyNotFound {
		v2, e2, _ := s.g.Do(key, func() (interface{}, error) {
			val, err := s.LoadFunc(ctx, key)
			if err == nil {
				err = s.Cache.Set(ctx, key, val, s.expiration)
				if err != nil {
					return val, fmt.Errorf("%w, reason: %s", cache.NewErrRefreshCacheFail(key), err.Error())
				}
			}
			return val, err
		})
		return v2.([]byte), e2
	}
	return v1, e1
}
