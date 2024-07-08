package local_cache

import (
	"context"
	"fmt"
	"github.com/ac-zht/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBuildInMapCache_Get(t *testing.T) {
	testCase := []struct {
		name      string
		cache     func() *BuildInMapCache
		key       string
		wantVal   any
		wantError error
	}{
		{
			name: "get",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10)
				_ = res.Set(context.Background(), "key", "value", time.Minute)
				return res
			},
			key:     "key",
			wantVal: "value",
		},
		{
			name: "not exist",
			cache: func() *BuildInMapCache {
				return NewBuildInMapCache(10)
			},
			key:       "key",
			wantError: fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, "key"),
		},
		{
			name: "expired",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10)
				_ = res.Set(context.Background(), "key", "value", -time.Minute)
				return res
			},
			key:       "key",
			wantError: fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, "key"),
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			val, err := cache.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func TestBuildInMapCache_IntervalEliminate(t *testing.T) {
	cnt := 0
	cache := NewBuildInMapCache(10, BuildInMapCacheWithOutInterval(time.Second), BuildInMapCacheWithEvictedCallback(func(key string, val any) {
		cnt++
	}))
	err := cache.Set(context.Background(), "k1", "v1", time.Second)
	err = cache.Set(context.Background(), "k2", "v2", time.Second*2)
	err = cache.Set(context.Background(), "k3", "v3", time.Second*3)
	assert.NoError(t, err)
	time.Sleep(time.Second * 4)
	_, ok := cache.Data["k3"]
	require.False(t, ok)
	require.Equal(t, 3, cnt)
}
