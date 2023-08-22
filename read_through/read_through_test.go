package read_through

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zht-account/cache"
	"testing"
	"time"
)

func TestReadThroughCache_Get(t *testing.T) {
	testCase := []struct {
		name      string
		cache     func() *ReadThroughCache
		key       string
		wantVal   []byte
		wantData  map[string][]byte
		wantError error
	}{
		{
			name: "cache exist",
			cache: func() *ReadThroughCache {
				return NewReadThroughCache(&MockCache{
					data: map[string][]byte{"k1": []byte("v1")},
				},
					time.Minute,
					func(ctx context.Context, key string) ([]byte, error) { return nil, errors.New("not found") },
				)
			},
			key:      "k1",
			wantVal:  []byte("v1"),
			wantData: map[string][]byte{"k1": []byte("v1")},
		},
		{
			name: "set cache success",
			cache: func() *ReadThroughCache {
				return NewReadThroughCache(&MockCache{
					data: make(map[string][]byte),
				},
					time.Minute,
					func(ctx context.Context, key string) ([]byte, error) { return []byte("v1"), nil },
				)
			},
			key:      "k1",
			wantVal:  []byte("v1"),
			wantData: map[string][]byte{"k1": []byte("v1")},
		},
		{
			name: "set cache fail",
			cache: func() *ReadThroughCache {
				return NewReadThroughCache(&MockCache{
					data:       make(map[string][]byte),
					setErrFlag: true,
				},
					time.Minute,
					func(ctx context.Context, key string) ([]byte, error) { return []byte("v1"), nil },
				)
			},
			key:       "k1",
			wantError: cache.NewErrRefreshCacheFail("k1"),
			wantData:  map[string][]byte{},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.cache()
			val, err := c.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantError, err)
			assert.Equal(t, tc.wantData, c.Cache.(*MockCache).data)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
}
func TestReadThroughCache_AsyncGet(t *testing.T) {
	testCase := []struct {
		name      string
		cache     func() *ReadThroughCache
		key       string
		wantVal   []byte
		wantData  map[string][]byte
		wantError error
	}{
		{
			name: "cache exist",
			cache: func() *ReadThroughCache {
				return NewReadThroughCache(&MockCache{
					data: map[string][]byte{"k1": []byte("v1")},
				},
					time.Minute,
					func(ctx context.Context, key string) ([]byte, error) { return nil, errors.New("not found") },
				)
			},
			key:      "k1",
			wantVal:  []byte("v1"),
			wantData: map[string][]byte{"k1": []byte("v1")},
		},
		{
			name: "cache not exist and set cache success",
			cache: func() *ReadThroughCache {
				return NewReadThroughCache(&MockCache{
					data: make(map[string][]byte),
				},
					time.Minute,
					func(ctx context.Context, key string) ([]byte, error) { return []byte("v1"), nil },
				)
			},
			key:       "k1",
			wantError: cache.ErrKeyNotFound,
		},
		{
			name: "cache not exist and set cache fail",
			cache: func() *ReadThroughCache {
				return NewReadThroughCache(&MockCache{
					data:       make(map[string][]byte),
					setErrFlag: true,
				},
					time.Minute,
					func(ctx context.Context, key string) ([]byte, error) { return nil, errors.New("not found") },
				)
			},
			key:       "k1",
			wantError: cache.ErrKeyNotFound,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.cache()
			val, err := c.AsyncGet(context.Background(), tc.key)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantData, c.Cache.(*MockCache).data)
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

type MockCache struct {
	cache.Cache
	data       map[string][]byte
	setErrFlag bool
}

func (c *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := c.data[key]
	if !ok {
		return nil, cache.ErrKeyNotFound
	}
	return val, nil
}

func (c *MockCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	if c.setErrFlag {
		return fmt.Errorf("set error: key : %s", key)
	}
	c.data[key] = val
	return nil
}

func (c *MockCache) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}
