package lark

import (
	"context"
	"github.com/Xhofe/go-cache"
	"time"
)

type TokenCache struct {
	cache.ICache[string]
}

func (t *TokenCache) Set(_ context.Context, key string, value string, expireTime time.Duration) error {
	t.ICache.Set(key, value, cache.WithEx[string](expireTime))

	return nil
}

func (t *TokenCache) Get(_ context.Context, key string) (string, error) {
	v, ok := t.ICache.Get(key)
	if ok {
		return v, nil
	}

	return "", nil
}

func newTokenCache() *TokenCache {
	c := cache.NewMemCache[string]()

	return &TokenCache{c}
}
