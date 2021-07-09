package main

import (
	"context"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/malyusha/immune-mosru-server/api/http/router/immune"
	"github.com/malyusha/immune-mosru-server/api/telegram"
	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/internal/bot"
	botRedis "github.com/malyusha/immune-mosru-server/internal/bot/redis"
	"github.com/malyusha/immune-mosru-server/internal/storage/inmem"
	"github.com/malyusha/immune-mosru-server/internal/storage/mongo"
	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
	"github.com/malyusha/immune-mosru-server/pkg/contextutils"
	inmemDB "github.com/malyusha/immune-mosru-server/pkg/database/inmem"
	"github.com/malyusha/immune-mosru-server/pkg/database/mongodb"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	metrics "github.com/malyusha/immune-mosru-server/pkg/metrics/prometheus"
	"github.com/malyusha/immune-mosru-server/pkg/redis"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/httputils"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/middleware"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/router"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/status"
	"github.com/malyusha/immune-mosru-server/pkg/waiter"
)

var (
	Version = "v0.1.0"
	Name    = "bot"
)

func main() {
	ctx := contextutils.WithSignals(context.Background())
	cfg, err := parseConfig()
	if err != nil {
		logger.Fatalf("failed to initialize cfg: %s", err)
	}

	log := configureLogger(cfg.Log)
	w := waiter.New(ctx)

	var certsStorage internal.CertificatesStorage
	if cfg.Storage.Type == "mongo" && cfg.Storage.Mongo.URI != "" {
		client, err := mongodb.NewClient(ctx, cfg.Storage.Mongo)
		if err != nil {
			logger.Fatalf("failed to create mongo connection: %s", err)
		}

		certsStorage, err = mongo.NewCertsStorage(ctx, mongodb.CollectionConfig{
			DB: client.Database("immune"),
		})
		if err != nil {
			logger.Fatalf("failed to create certs storage: %s", err)
		}
	} else {
		db := inmemDB.New()
		certsStorage = inmem.NewCertificatesStorage(db)
	}
	var botDataStorage telegram.DataStorage = nil
	var botStateManager telegram.StateManager = nil
	redis.SetPrefix(Name)
	// Init redis storage for data and state of bot if redis is configured.
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatalf("failed to initialize redis client: %s", err)
	}
	cache := redis.NewCache(redisClient)

	botDataStorage = botRedis.NewRedisDataStorage(cache)
	botStateManager = botRedis.NewStateStorage(cache)

	// Create telegram bot
	vaxService := vaxcert.NewService(certsStorage, cache)
	qrGen := bot.NewQRGenerator(cfg.QR.URLPattern)

	commonMetrics := metrics.New(prometheus.Labels{"app": Name, "version": Version})

	immuneRouter := immune.NewRouter(vaxService)

	// Registering global router for application handler
	routerHandler, err := router.NewRouterHandler(
		router.WithMetrics(commonMetrics.HTTPHandler),
		router.WithRouter(immuneRouter),
		router.WithGlobalMiddleware(
			middleware.PassTraceIDToContext(contextutils.NewTraceID),
			middleware.PassLoggingFieldsToRequestContext(httputils.ExtractLogFields),
			middleware.SetContentType("application/json"),
			middleware.RecoverOnPanic(log),
			middleware.LogRequests(
				middleware.IgnoreHeaders("*"),
			),
		),
	)

	if err != nil {
		log.Fatalf("failed to initialize router handler: %s", err)
	}

	appServer := server.New(
		server.WithName("app"),
		server.WithHandler(routerHandler),
		server.WithAddr(cfg.HTTPServer.Addr),
		server.WithTimeouts(serverTimeouts(cfg.HTTPServer)),
	)

	statusServer, ready := createStatusServer(cfg.HTTPStatusServer)

	tgBotOts := []telegram.Option{
		telegram.WithCertificatesService(vaxService),
		telegram.WithQRGenerator(qrGen),
		telegram.WithDataStorage(botDataStorage),
		telegram.WithStateManager(botStateManager),
	}
	if cfg.Telegram.WebhookURL != "" && cfg.Telegram.ListenAddr != "" {
		tgBotOts = append(tgBotOts, telegram.WithWebhook(cfg.Telegram.WebhookURL, cfg.Telegram.ListenAddr))
	}
	tgBot, err := telegram.NewBot(cfg.Telegram.Token, tgBotOts...)

	if err != nil {
		logger.Fatalf("failed to run telegram bot: %s", err)
	}

	w.Add(tgBot, appServer, statusServer)

	ready()
	if err := w.Run(); err != nil {
		log.Fatalf("error running server: %s", err)
	}

	log.Info("exiting")
}

// configures global logger instance.
func configureLogger(cfg *LogConfig) logger.Logger {
	err := logger.Configure(&logger.Config{
		Raw:    cfg.Raw,
		Output: cfg.Output,
		Level:  cfg.Level,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure logger: %s\n", err)
		os.Exit(1)
	}

	return logger.GetLogger()
}

func serverTimeouts(cfg *HTTPServerConfig) server.Timeouts {
	return server.Timeouts{
		Write:    cfg.WriteTimeout.Duration,
		Read:     cfg.ReadTimeout.Duration,
		Idle:     cfg.IdleTimeout.Duration,
		Shutdown: cfg.ShutdownTimeout.Duration,
	}
}

func createStatusServer(cfg *HTTPServerConfig) (*server.Server, func()) {
	readyCh := make(chan struct{}, 1)
	ready := func() {
		readyCh <- struct{}{}
	}

	return server.New(
		server.WithName("status"),
		server.WithAddr(cfg.Addr),
		server.WithHandler(status.NewHandler(readyCh)),
		server.WithTimeouts(serverTimeouts(cfg)),
	), ready
}
