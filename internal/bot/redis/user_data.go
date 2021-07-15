package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/cache/v8"

	"github.com/malyusha/immune-mosru-server/api/telegram"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/redis"
)

type dataStorage struct {
	cache  *redis.Cache
	logger logger.Logger
}

func (r *dataStorage) GetData(ctx context.Context, id string) (*telegram.UserData, error) {
	dst := new(telegram.UserData)
	log := r.logger.WithContext(ctx)
	err := r.cache.Get(ctx, getDataByUserIDKey(id), dst)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return nil, telegram.ErrDataMissing
		}

		return nil, fmt.Errorf("failed to get data from cache: %w", err)
	}

	log.Debugf("received data for user ID %s", id)

	return dst, nil
}

func (r *dataStorage) SetData(ctx context.Context, id string, data *telegram.UserData) error {
	log := r.logger.WithContext(ctx)
	item := &cache.Item{
		Ctx:   ctx,
		Key:   getDataByUserIDKey(id),
		Value: data,
	}
	if err := r.cache.Set(item); err != nil {
		return fmt.Errorf("failed to set data for id %d: %w", id, err)
	}

	log.Debugf("new data is written for user ID %s", id)

	return nil
}

func NewRedisDataStorage(cache *redis.Cache) *dataStorage {
	return &dataStorage{
		cache:  cache,
		logger: logger.With(logger.Fields{"package": "bot-data-storage-redis"}),
	}
}

func getDataByUserIDKey(id string) string {
	return fmt.Sprintf("%s:%s", userDataKeyPrefix, id)
}

var (
	userDataKeyPrefix = redis.WithCachePrefix("bot_user_data")
)
