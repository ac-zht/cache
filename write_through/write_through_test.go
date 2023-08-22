package write_through

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zht-account/cache"
	"testing"
	"time"
)

func TestWriteThroughCache_Get(t *testing.T) {
	testCase := []struct {
		name      string
		cache     func() *WriteThroughCache
		key       string
		val       []byte
		wantData  map[string][]byte
		wantError error
	}{
		{
			name: "success",
			cache: func() *WriteThroughCache {
				return NewWriteThroughCache(&MockCache{
					data: map[string][]byte{"k1": []byte("v1")},
				},
					func(ctx context.Context, key string, val []byte) error { return nil },
				)
			},
			key:      "k1",
			val:      []byte("v1"),
			wantData: map[string][]byte{"k1": []byte("v1")},
		},
		{
			name: "fail",
			cache: func() *WriteThroughCache {
				return NewWriteThroughCache(&MockCache{
					data: map[string][]byte{"k1": []byte("v1")},
				},
					func(ctx context.Context, key string, val []byte) error { return errors.New("database store fail") },
				)
			},
			key:       "k1",
			val:       []byte("v1"),
			wantError: errors.New("database store fail"),
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.cache()
			err := c.Set(context.Background(), tc.key, tc.val, time.Minute)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantData, c.Cache.(*MockCache).data)
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
