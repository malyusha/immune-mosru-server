package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	redisPkg "github.com/go-redis/redis/v8"
	"github.com/ilyakaznacheev/cleanenv"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/malyusha/immune-mosru-server/pkg/database/mongodb"
	"github.com/malyusha/immune-mosru-server/pkg/fs"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/redis"
)

// Config is the main structure for application configuration
type Config struct {
	// Server configuration
	AppHTTPServer *AppHTTPConfig `yaml:"app_http,omitempty" env-prefix:"APP_SERVER_"`
	// Server configuration
	StatusHTTPServer *StatusHTTPConfig `yaml:"status_http,omitempty" env-prefix:"STATUS_SERVER_"`
	// Configuration of telegram bot.
	Telegram TGBotConfig `yaml:"telegram" env-prefix:"TELEGRAM_"`
	// QR generator config
	QR QRConfig `yaml:"qr"`
	// Redis client configuration
	Redis redis.Config `yaml:"redis"`
	// Mongo is the mongo client configuration.
	Mongo mongodb.Config `yaml:"mongo"`
	// Log configuration
	Log *LogConfig `yaml:"log"`
}

// initializeRedisClient initializes new redis client and returns it.
// If client initialization fails then fatal is logged and process interrupts.
func (c *Config) initializeRedisClient(ctx context.Context) redisPkg.Cmdable {
	redisClient, err := redis.NewClient(c.Redis)
	if err != nil {
		logger.Fatalf("failed to initialize redis client: %s", err)
	}

	return redisClient
}

// initializeMongoClient returns newly initialized mongoDB client with provided configuration.
// If client creation fails then fatal is logged and process interrupts.
func (c *Config) initializeMongoClient(ctx context.Context) *mongo.Client {
	client, err := mongodb.NewClient(ctx, c.Mongo)
	if err != nil {
		logger.Fatalf("failed to create mongo connection: %s", err)
	}

	return client
}

func (c *Config) configureLogger() logger.Logger {
	if c.Log == nil {
		return logger.GetLogger()
	}
	err := logger.Configure(&logger.Config{
		Raw:   c.Log.Raw,
		Level: c.Log.Level,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure logger: %s\n", err)
		os.Exit(1)
	}

	return logger.GetLogger()
}

type StatusHTTPConfig struct {
	HTTPServerConfig
	Addr string `yaml:"addr" env:"ADDR" env-default:"9090"`
}

type AppHTTPConfig struct {
	HTTPServerConfig
	Addr string `yaml:"addr" env:"ADDR" env-default:":8080"`
}

type HTTPServerConfig struct {
	// Duration in seconds for which the server will wait existing connections to finish.
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout,omitempty" env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"15s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"15s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"15s"`
}

// QRConfig represents QR generator configuration.
type QRConfig struct {
	// URLPattern is the string URL with %s template placed.
	URLPattern string `yaml:"url_pattern" env:"QR_URL_PATTERN"`
}

// LogConfig represents logger configuration.
type LogConfig struct {
	Level string `yaml:"level" env:"LOG_LEVEL"`
	Raw   bool   `yaml:"raw" env:"LOG_RAW"`
}

// TGBotConfig represents telegram bot configuration.
type TGBotConfig struct {
	Token      string `yaml:"token" env:"BOT_TOKEN"`
	WebhookURL string `yaml:"webhook_url" env:"WEBHOOK_URL"`
	ListenAddr string `yaml:"webhook_listen_addr" env:"WEBHOOK_LISTEN_ADDR"`
}

// LoadConfig returns new config struct from config file path
func LoadConfig() (*Config, error) {
	fset := flag.NewFlagSet(Name, flag.ContinueOnError)
	configPath := fset.String("config", "", "configuration file path")

	var cfg Config
	fset.Usage = cleanenv.FUsage(fset.Output(), &cfg, nil, fset.Usage)
	if err := fset.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	var err error
	if configPath == nil || !fs.FileExists(*configPath) {
		err = cleanenv.ReadEnv(&cfg)
	} else {
		err = cleanenv.ReadConfig(*configPath, &cfg)
	}

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// isRedisEnabled checks whether redis addr is passed into configuration.
func (c *Config) isRedisEnabled() bool {
	return c.Redis.Addr != ""
}
