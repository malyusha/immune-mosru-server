package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

func NewClient(cfg Config) (r redis.Cmdable, err error) {
	switch cfg.Mode {
	case configModeSingle:
		r, err = newSingleClient(&cfg)
	case configModeCluster:
		r, err = newClusterClient(&cfg)
	default:
		return nil, errors.New(fmt.Sprintf("unknown redis mode %s", cfg.Mode))
	}

	if err != nil {
		return nil, err
	}

	err = r.Ping(context.Background()).Err()

	return
}

func newSingleClient(cfg *Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.Addr)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opt), nil
}

func newClusterClient(cfg *Config) (*redis.ClusterClient, error) {
	addrStrings := strings.Split(strings.Replace(cfg.Addr, " ", "", -1), ",")
	addrs := make([]string, 0, len(addrStrings))
	for _, a := range addrStrings {
		opt, err := redis.ParseURL(a)
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, opt.Addr)
	}

	redisConfig := &redis.ClusterOptions{
		Addrs: addrs,
	}

	return redis.NewClusterClient(redisConfig), nil
}
