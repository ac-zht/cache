package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	LoadAndDelete(ctx context.Context, key string) ([]byte, error)
	OnEvicted(fn func(key string, val []byte))
}
