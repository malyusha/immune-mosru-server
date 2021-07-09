package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/malyusha/immune-mosru-server/pkg/database/mongodb"
	"github.com/malyusha/immune-mosru-server/pkg/redis"
)

// Duration is the custom duration
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(b []byte) error {
	if d != nil {
		return nil
	}
	parsed, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}

	d.Duration = parsed

	return nil
}

// Config is the main structure for application configuration
type Config struct {
	// Server configuration
	HTTPServer *HTTPServerConfig `json:"http,omitempty"`

	// Server configuration
	HTTPStatusServer *HTTPServerConfig `json:"status_http,omitempty"`
	// Configuration of telegram bot.
	Telegram TGBotConfig `json:"telegram"`
	// QR generator config
	QR QRConfig `json:"qr"`
	// Redis client configuration
	Redis   redis.Config  `json:"redis"`
	Storage StorageConfig `json:"storage"`
	// Log configuration
	Log *LogConfig `json:"log"`
}

type HTTPServerConfig struct {
	// HTTP server address.
	Addr string `json:"addr"`

	// Duration in seconds for which the server will wait existing connections to finish.
	ShutdownTimeout Duration `json:"shutdown_timeout,omitempty"`

	// See defaults.go
	ReadTimeout  Duration `json:"read_timeout"`
	WriteTimeout Duration `json:"write_timeout"`
	IdleTimeout  Duration `json:"idle_timeout"`
}

type StorageConfig struct {
	Type  string         `json:"type"`
	Mongo mongodb.Config `json:"mongo"`
}

type QRConfig struct {
	URLPattern string `json:"url_pattern"`
}

type LogConfig struct {
	Level  string    `json:"level"`
	Output io.Writer `json:"-"`
	Raw    bool      `json:"raw"`
}

// TGBotConfig represents telegram bot configuration.
type TGBotConfig struct {
	Token      string `json:"token"`
	WebhookURL string `json:"webhook_url"`
	ListenAddr string `json:"listen_addr"`
}

// ReadFromEnv reads some configuration parameters from ENV variables.
func (c *Config) ReadFromEnv() {
	c.HTTPServer.Addr = envOrDefault("HTTP_ADDR", c.HTTPServer.Addr)
	c.HTTPStatusServer.Addr = envOrDefault("STATUS_HTTP_ADDR", c.HTTPStatusServer.Addr)

	redisMode := redis.ConfigMode(envOrDefault("REDIS_MODE", string(c.Redis.Mode)))
	c.Redis.Mode = redisMode
	c.Redis.Addr = envOrDefault("REDIS_ADDR", c.Redis.Addr)

	c.Log.Level = envOrDefault("LOG_LEVEL", c.Log.Level)
	c.Log.Raw = envBoolOrDefault("LOG_RAW", c.Log.Raw)
	c.Storage.Mongo.URI = envOrDefault("MONGO_URI", c.Storage.Mongo.URI)

	c.QR.URLPattern = envOrDefault("QR_URL_PATTERN", c.QR.URLPattern)

	c.Telegram.Token = envOrDefault("TELEGRAM_BOT_TOKEN", c.Telegram.Token)
	c.Telegram.WebhookURL = envOrDefault("TELEGRAM_WEBHOOK_URL", c.Telegram.WebhookURL)
	c.Telegram.ListenAddr = envOrDefault("TELEGRAM_LISTEN_ADDR", c.Telegram.ListenAddr)
}

// DefaultConfig returns config instance with default settings
func DefaultConfig() *Config {
	return &Config{
		HTTPServer: &HTTPServerConfig{
			Addr: DefaultServerAddr,
		},
		HTTPStatusServer: &HTTPServerConfig{
			Addr: DefaultStatusServerAddr,
		},
		Log: &LogConfig{
			Output: os.Stderr,
		},
		Redis: redis.Config{
			Addr: "redis://127.0.0.1:6379",
		},
		Storage: StorageConfig{
		},
	}
}

// LoadConfig returns new config struct from config file path
func LoadConfig(configPath, valuesPath string) (*Config, error) {
	values := make(map[string]string)
	// Initialize base config
	cfg := DefaultConfig()
	if fileExists(valuesPath) {
		valuesData, err := ioutil.ReadFile(valuesPath)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := json.Unmarshal(valuesData, &values); err != nil {
			return nil, fmt.Errorf("unmarshal values: %w", err)
		}
	}

	if fileExists(configPath) {
		t, err := template.New(filepath.Base(configPath)).ParseFiles(configPath)
		if err != nil {
			return nil, fmt.Errorf("parse config template: %w", err)
		}
		t.Option("missingkey=zero")

		var buf bytes.Buffer
		if err := t.Execute(&buf, values); err != nil {
			return nil, fmt.Errorf("execute config template: %w", err)
		}
		configData := buf.Bytes()

		// Unmarshal config file into default config
		if err := json.Unmarshal(configData, cfg); err != nil {
			return nil, err
		}
	}

	// Reading params from ENV if they've been passed
	cfg.ReadFromEnv()

	return cfg, nil
}

func parseConfig() (*Config, error) {
	var configFile string
	var valuesFile string
	flag.StringVar(&configFile, "config", "configs/config.json", "Path to the configuration file")
	flag.StringVar(&valuesFile, "values", "configs/values.json", "Path to the values file")
	flag.Parse()

	cfg, err := LoadConfig(configFile, valuesFile)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	file, err := os.Open(path)

	return !os.IsNotExist(err) && file != nil
}

func envBoolOrDefault(key string, def bool) bool {
	if _, ok := os.LookupEnv(key); ok {
		return true
	}

	return def
}

func envOrDefault(key string, def string) string {
	if envVal := os.Getenv(key); envVal != "" {
		return envVal
	}

	return def
}

func (c *Config) isRedisEnabled() bool {
	return c.Redis.Addr != ""
}
