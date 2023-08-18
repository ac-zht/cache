package cache

import (
	"context"
	"errors"
	"github.com/zht-account/gotools/list"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache: key not exist")
)

type MaxMemoryCache struct {
	Cache
	max  int64
	used int64

	keys  *list.LinkedList[string]
	mutex *sync.Mutex
}

func NewMaxMemoryCache(max int64, cache Cache) *MaxMemoryCache {
	res := &MaxMemoryCache{
		Cache: cache,
		max:   max,
		keys:  list.NewLinkedList[string](),
		mutex: &sync.Mutex{},
	}
	res.Cache.OnEvicted(res.evicted)
	return res
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	//为了保证keys中key淘汰顺序
	_, _ = m.Cache.LoadAndDelete(ctx, key)
	for m.used+int64(len(val)) > m.max {
		expKey, err := m.keys.Get(0)
		if err != nil {
			return err
		}
		err = m.Cache.Delete(ctx, expKey)
		if err != nil {
			return err
		}
	}
	err := m.Cache.Set(ctx, key, val, expiration)
	if err == nil {
		m.used += int64(len(val))
		_ = m.keys.Append(key)
	}
	return err
}

func (m *MaxMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, err := m.Cache.Get(ctx, key)
	if err == nil {
		m.deleteKey(key)
		_ = m.keys.Append(key)
		return val, nil
	}
	return val, errKeyNotFound
}

func (m *MaxMemoryCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Cache.Delete(ctx, key)
}

func (m *MaxMemoryCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Cache.LoadAndDelete(ctx, key)
}

func (m *MaxMemoryCache) OnEvicted(fn func(key string, val []byte)) {
	m.Cache.OnEvicted(func(key string, val []byte) {
		m.evicted(key, val)
		fn(key, val)
	})
}

func (m *MaxMemoryCache) evicted(key string, val []byte) {
	m.used -= int64(len(val))
	m.deleteKey(key)
}

func (m *MaxMemoryCache) deleteKey(key string) {
	keys := m.keys.AsSlice()
	for i, val := range keys {
		if val == key {
			_, _ = m.keys.Delete(i)
			return
		}
	}
}
