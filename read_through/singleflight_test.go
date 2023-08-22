package read_through

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestSingleflightCache_Get(t *testing.T) {
	loadFunc := func(ctx context.Context, key string) ([]byte, error) {
		return []byte("value"), nil
	}
	cache := NewSingleflightCache(&MockCacheV2{MockCache{data: map[string][]byte{}}}, time.Minute, loadFunc)
	wg := &sync.WaitGroup{}
	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val, _ := cache.Get(context.Background(), "key")
			values[i] = string(val)
		}(i)
	}
	wg.Wait()
	all := true
	for _, v := range values {
		if v != "value" {
			all = false
		}
	}
	assert.True(t, all)
}

type MockCacheV2 struct {
	MockCache
}

func (c *MockCacheV2) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	return nil
}
