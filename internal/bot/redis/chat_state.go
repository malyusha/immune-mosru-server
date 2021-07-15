package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/cache/v8"

	"github.com/malyusha/immune-mosru-server/api/telegram"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/redis"
)

type stateStorage struct {
	cache  *redis.Cache
	logger logger.Logger
}

func (s *stateStorage) GetState(ctx context.Context, id string) (telegram.ChatState, error) {
	dst := new(telegram.ChatState)
	log := s.logger.WithContext(ctx)
	err := s.cache.Get(ctx, getStateByUserIDKey(id), dst)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return "", telegram.ErrStateMissing
		}

		return "", fmt.Errorf("failed to get data from cache: %w", err)
	}

	log.Debugf("received state for user ID %s: %s", id, *dst)

	return *dst, nil
}

func (s *stateStorage) SetState(ctx context.Context, id string, state telegram.ChatState) error {
	log := s.logger.WithContext(ctx)
	item := &cache.Item{
		Ctx:   ctx,
		Key:   getStateByUserIDKey(id),
		Value: state,
	}

	if err := s.cache.Set(item); err != nil {
		return fmt.Errorf("failed to write state for id %d: %w", id, err)
	}

	log.Debugf("new state is written for user ID %s", id)

	return nil
}

func getStateByUserIDKey(id string) string {
	return fmt.Sprintf("%s:%s", userStateKeyPrefix, id)
}

func NewStateStorage(cache *redis.Cache) *stateStorage {
	return &stateStorage{
		cache:  cache,
		logger: logger.With(logger.Fields{"package": "bot-state-storage-redis"}),
	}
}

var (
	userStateKeyPrefix = redis.WithCachePrefix("bot_user_state")
)
