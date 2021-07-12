package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

type Config struct {
	Addr string `yaml:"addr" env:"MONGO_ADDR"`
}

// NewClient returns new client wrapper for mongo client and database.
func NewClient(ctx context.Context, cfg Config) (*mongo.Client, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.Addr))
	if err != nil {
		return nil, fmt.Errorf("create client error: %s", err)
	}

	// trying to connect
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connection error: %s", err)
	}

	logger.Debug("successfully connected to database. ping")
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("ping database error: %s", err)
	}

	logger.Debugf("database is alive")
	return client, nil
}

func validateConfig(cfg Config) error {
	if cfg.Addr == "" {
		return errors.New("mongo connection Addr not provided")
	}

	return nil
}
