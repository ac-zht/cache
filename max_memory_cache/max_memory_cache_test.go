package max_memory_cache

import (
	"context"
	"errors"
	"github.com/ac-zht/cache"
	"github.com/ac-zht/gotools/list"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMaxMemoryCache_Set(t *testing.T) {
	testCase := []struct {
		name      string
		m         func() *MaxMemoryCache
		key       string
		value     []byte
		wantUsed  int64
		wantKeys  []string
		wantError error
	}{
		//不存在key，直接加入
		{
			name: "not exist",
			m: func() *MaxMemoryCache {
				return NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{},
				})
			},
			key:      "key",
			value:    []byte("value"),
			wantUsed: 5,
			wantKeys: []string{"key"},
		},
		//已存在key，重新设置增加内存量
		{
			name: "existed and add used",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{
						"k":   []byte("v"),
						"key": []byte("value"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k", "key"})
				return res
			},
			key:      "key",
			value:    []byte("this value"),
			wantUsed: 11,
			wantKeys: []string{"k", "key"},
		},
		//已存在key，重新设置减少内存量
		{
			name: "existed and reduce used",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{
						"k":   []byte("v"),
						"key": []byte("value"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k", "key"})
				return res
			},
			key:      "key",
			value:    []byte("v"),
			wantUsed: 2,
			wantKeys: []string{"k", "key"},
		},
		//内存已满，淘汰一次
		{
			name: "delete one times",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{
					data: map[string][]byte{
						"key": []byte("value"),
					},
				})
				res.used = 5
				res.keys = list.NewLinkedListOf[string]([]string{"key"})
				return res
			},
			key:      "this key",
			value:    []byte("this value"),
			wantUsed: 10,
			wantKeys: []string{"this key"},
		},
		//内存已满，淘汰多次
		{
			name: "delete multi times",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(20, &mockCache{
					data: map[string][]byte{
						"k1": []byte("v1"),
						"k2": []byte("v2"),
						"k3": []byte("v3"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k1", "k2", "k3"})
				return res
			},
			key:      "k4",
			value:    []byte("this is a new value"),
			wantUsed: 19,
			wantKeys: []string{"k4"},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.m()
			err := m.Set(context.Background(), tc.key, tc.value, time.Minute)
			assert.Equal(t, tc.wantError, err)
			assert.Equal(t, tc.wantKeys, m.keys.AsSlice())
			assert.Equal(t, tc.wantUsed, m.used)
		})
	}
}

func TestMaxMemoryCache_Get(t *testing.T) {
	testCase := []struct {
		name      string
		m         func() *MaxMemoryCache
		key       string
		wantVal   []byte
		wantKeys  []string
		wantError error
	}{
		{
			name: "not exist",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{
					data: map[string][]byte{
						"k1": []byte("v1"),
						"k2": []byte("v2"),
						"k3": []byte("v3"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k1", "k2", "k3"})
				return res
			},
			key:       "k4",
			wantError: cache.ErrKeyNotFound,
		},
		{
			name: "exist",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{
					data: map[string][]byte{
						"k1": []byte("v1"),
						"k2": []byte("v2"),
						"k3": []byte("v3"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k1", "k2", "k3"})
				return res
			},
			key:      "k1",
			wantVal:  []byte("v1"),
			wantKeys: []string{"k2", "k3", "k1"},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.m()
			val, err := m.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
			assert.Equal(t, tc.wantKeys, m.keys.AsSlice())
		})
	}
}

func TestMaxMemoryCache_Delete(t *testing.T) {
	testCase := []struct {
		name      string
		m         func() *MaxMemoryCache
		key       string
		wantUsed  int64
		wantKeys  []string
		wantError error
	}{
		{
			name: "no delete",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{
					data: map[string][]byte{
						"k1": []byte("v1"),
						"k2": []byte("v2"),
						"k3": []byte("v3"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k1", "k2", "k3"})
				return res
			},
			key:      "k4",
			wantUsed: 6,
			wantKeys: []string{"k1", "k2", "k3"},
		},
		{
			name: "deleted",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{
					data: map[string][]byte{
						"k1": []byte("v1"),
						"k2": []byte("v2"),
						"k3": []byte("v3"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k1", "k2", "k3"})
				return res
			},
			key:      "k2",
			wantUsed: 4,
			wantKeys: []string{"k1", "k3"},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.m()
			err := m.Delete(context.Background(), tc.key)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantUsed, m.used)
			assert.Equal(t, tc.wantKeys, m.keys.AsSlice())
		})
	}
}

func TestMaxMemoryCache_LoadAndDelete(t *testing.T) {
	testCase := []struct {
		name      string
		m         func() *MaxMemoryCache
		key       string
		wantVal   []byte
		wantUsed  int64
		wantKeys  []string
		wantError error
	}{
		{
			name: "no delete",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{
					data: map[string][]byte{
						"k1": []byte("v1"),
						"k2": []byte("v2"),
						"k3": []byte("v3"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k1", "k2", "k3"})
				return res
			},
			key:       "k4",
			wantUsed:  6,
			wantKeys:  []string{"k1", "k2", "k3"},
			wantError: cache.ErrKeyNotFound,
		},
		{
			name: "deleted",
			m: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(10, &mockCache{
					data: map[string][]byte{
						"k1": []byte("v1"),
						"k2": []byte("v2"),
						"k3": []byte("v3"),
					},
				})
				res.used = 6
				res.keys = list.NewLinkedListOf[string]([]string{"k1", "k2", "k3"})
				return res
			},
			key:      "k2",
			wantVal:  []byte("v2"),
			wantUsed: 4,
			wantKeys: []string{"k1", "k3"},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.m()
			val, err := m.LoadAndDelete(context.Background(), tc.key)
			assert.Equal(t, tc.wantError, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
			assert.Equal(t, tc.wantUsed, m.used)
			assert.Equal(t, tc.wantKeys, m.keys.AsSlice())
		})
	}
}

type mockCache struct {
	cache.Cache
	data map[string][]byte
	fn   func(key string, val []byte)
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, errors.New("not exist")
	}
	return val, nil
}

func (m *mockCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	m.data[key] = val
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	val, ok := m.data[key]
	if ok {
		delete(m.data, key)
		m.fn(key, val)
	}
	return nil
}

func (m *mockCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, cache.ErrKeyNotFound
	}
	delete(m.data, key)
	m.fn(key, val)
	return val, nil
}

func (m *mockCache) OnEvicted(fn func(key string, val []byte)) {
	m.fn = fn
}
