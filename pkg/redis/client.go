package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/malyusha/immune-mosru-server/pkg/logger"
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

	retries := 3
	for retries > 0 {
		err = r.Ping(context.Background()).Err()
		if err == nil {
			break
		}
		var nErr *net.OpError
		if errors.As(err, &nErr) {
			logger.Error("failed to connect to redis. retrying in 1 second")
			time.Sleep(time.Second)
			retries--
			continue
		}

		return nil, err
	}

	return
}

func newSingleClient(cfg *Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.Addr)
	if err != nil {
		return nil, err
	}

	// just to be sure, that redis has started when initialized first time inside docker compose
	opt.MaxRetryBackoff = time.Second

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
