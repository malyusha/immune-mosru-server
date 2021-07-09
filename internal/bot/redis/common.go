package redis

import (
	"github.com/malyusha/immune-mosru-server/pkg/redis"
)

func init() {
	redis.SetPrefix("inmem")
}
