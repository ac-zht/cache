package cache

import (
    "context"
    "fmt"
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

func TestMaxCntCache_Set(t *testing.T) {
    testCase := []struct {
        name      string
        cache     func() *MaxCntCache
        key       string
        val       any
        wantCnt   int
        wantError error
    }{
        {
            name: "set and add cnt",
            cache: func() *MaxCntCache {
                res := NewMaxCntCache(10, NewBuildInMapCache(10), MaxCntCacheWithEvictedCallback())
                return res
            },
            key:     "key",
            val:     "value",
            wantCnt: 1,
        },
        {
            name: "set and cnt unchanged",
            cache: func() *MaxCntCache {
                res := NewMaxCntCache(10, NewBuildInMapCache(10), MaxCntCacheWithEvictedCallback())
                _ = res.Set(context.Background(), "key", "value", time.Minute)
                return res
            },
            key:     "key",
            val:     "value",
            wantCnt: 1,
        },
        {
            name: "set fail",
            cache: func() *MaxCntCache {
                res := NewMaxCntCache(2, NewBuildInMapCache(2), MaxCntCacheWithEvictedCallback())
                _ = res.Set(context.Background(), "k1", "v1", time.Minute)
                _ = res.Set(context.Background(), "k2", "v2", time.Minute)
                return res
            },
            key:       "k3",
            val:       "v3",
            wantError: fmt.Errorf("%w, key : %s", errKeyExceedMaxCnt, "k3"),
        },
    }
    for _, tc := range testCase {
        t.Run(tc.name, func(t *testing.T) {
            cache := tc.cache()
            err := cache.Set(context.Background(), tc.key, tc.val, time.Minute)
            assert.Equal(t, tc.wantError, err)
            if err != nil {
                return
            }
            assert.Equal(t, tc.wantCnt, cache.cnt)
        })
    }
}

func TestMaxCntCache_LoadAndDelete(t *testing.T) {
    testCase := []struct {
        name      string
        cache     func() *MaxCntCache
        key       string
        wantCnt   int
        wantVal   any
        wantError error
    }{
        {
            name: "deleted",
            cache: func() *MaxCntCache {
                res := NewMaxCntCache(10, NewBuildInMapCache(10), MaxCntCacheWithEvictedCallback())
                _ = res.Set(context.Background(), "key", "value", time.Minute)
                return res
            },
            key:     "key",
            wantCnt: 0,
            wantVal: "value",
        },
        {
            name: "not exist",
            cache: func() *MaxCntCache {
                res := NewMaxCntCache(10, NewBuildInMapCache(10), MaxCntCacheWithEvictedCallback())
                _ = res.Set(context.Background(), "k1", "v1", time.Minute)
                return res
            },
            key:       "k2",
            wantError: fmt.Errorf("%w, key: %s", errKeyNotFound, "k2"),
        },
    }
    for _, tc := range testCase {
        t.Run(tc.name, func(t *testing.T) {
            cache := tc.cache()
            val, err := cache.LoadAndDelete(context.Background(), tc.key)
            assert.Equal(t, tc.wantError, err)
            if err != nil {
                return
            }
            assert.Equal(t, tc.wantCnt, cache.cnt)
            assert.Equal(t, tc.wantVal, val)
        })
    }
}
