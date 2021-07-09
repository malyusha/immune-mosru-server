package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Cache struct {
	*cache.Cache
	Redis redis.Cmdable
}

var KeyPrefix = "app"

// sets global prefix
func SetPrefix(prefix string) {
	KeyPrefix = prefix
}

func NewCache(client redis.Cmdable) *Cache {
	cacherOpts := &cache.Options{
		Redis:      client,
		LocalCache: cache.NewTinyLFU(5000, time.Hour*24),
	}

	return &Cache{Redis: client, Cache: cache.New(cacherOpts)}
}

// DeleteMultiplePrefixes uses DeleteForPrefix func for each of key prefix in provided `keys` slice.
func DeleteMultiplePrefixes(ctx context.Context, cache *Cache, prefixes []string) error {
	eg, ctx := errgroup.WithContext(ctx)
	for _, key := range prefixes {
		key := key
		eg.Go(func() error {
			return DeleteForPrefix(ctx, cache, key)
		})
	}

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "delete multiple prefixes")
	}

	return nil
}

// DeleteForPrefix removes cache entries from redis using given key as prefix.
func DeleteForPrefix(ctx context.Context, cache *Cache, prefix string) error {
	iter := cache.Redis.Scan(ctx, 0, prefix+"*", 0).Iterator()
	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if len(keys) == 0 {
		return nil
	}

	err := cache.Redis.Del(ctx, keys...).Err()
	if err != nil && err != redis.Nil {
		return errors.Wrap(err, fmt.Sprintf("failed to delete keys %q with prefix %q", keys, prefix))
	}

	if iter.Err() != nil {
		return errors.Wrap(iter.Err(), fmt.Sprintf("failed to iterate over list keys using Scan for prefix with prefix %q", prefix))
	}

	return nil
}

// prefixes given key with default defaultPrefix
func WithCachePrefix(key string) string {
	return fmt.Sprintf("%s:%s", KeyPrefix, key)
}

func MakeHash(o interface{}) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%v", o)))
	return hex.EncodeToString(h[:])
}
