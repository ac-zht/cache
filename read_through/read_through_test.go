package read_through

import (
    "context"
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
        wantError error
    }{
        {
            name: "exist",
            cache: func() *ReadThroughCache {
                return NewReadThroughCache(&MockCache{},
                    time.Minute,
                    func(ctx context.Context, key string) ([]byte, error) { return []byte("value"), nil },
                    func(ctx context.Context, key string, val []byte) error { return nil },
                    func(ctx context.Context, key string) error { return nil },
                )
            },
            key: "k1",
        },
    }
    for _, tc := range testCase {
        t.Run(tc.name, func(t *testing.T) {

        })
    }
}

type MockCache struct {
    cache.Cache
}
