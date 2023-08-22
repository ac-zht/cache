package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrKeyNotFound = errors.New("cache: key not exist")
)

func NewErrKeyNotFound(key string) error {
	return fmt.Errorf("cache: key not exist, key : %s", key)
}

func NewErrRefreshCacheFail(key string) error {
	return fmt.Errorf("cache: refresh cache fail, key : %s", key)
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	LoadAndDelete(ctx context.Context, key string) ([]byte, error)
	OnEvicted(fn func(key string, val []byte))
}
