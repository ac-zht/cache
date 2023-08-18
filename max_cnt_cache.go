package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var errKeyExceedMaxCnt = errors.New("cache: key exceed max cnt")

type MaxCntCache struct {
	*BuildInMapCache
	max int
	cnt int
}

type MaxCntCacheOption func(cache *MaxCntCache)

func NewMaxCntCache(max int, cache *BuildInMapCache, opts ...MaxCntCacheOption) *MaxCntCache {
	res := &MaxCntCache{
		BuildInMapCache: cache,
		max:             max,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func (m *MaxCntCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.data[key]
	if !ok {
		if m.cnt+1 > m.max {
			return fmt.Errorf("%w, key : %s", errKeyExceedMaxCnt, key)
		}
		m.cnt++
	}
	return m.BuildInMapCache.Set(ctx, key, val, expiration)
}

func MaxCntCacheWithEvictedCallback() MaxCntCacheOption {
	return func(cache *MaxCntCache) {
		cache.onEvicted = func(key string, val any) {
			cache.cnt--
		}
	}
}
