package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/malyusha/immune-mosru-server/api/http/router/immune"
	"github.com/malyusha/immune-mosru-server/api/telegram"
	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/internal/bot"
	botRedis "github.com/malyusha/immune-mosru-server/internal/bot/redis"
	"github.com/malyusha/immune-mosru-server/internal/storage/mongo"
	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
	"github.com/malyusha/immune-mosru-server/pkg/contextutils"
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
	cfg, err := LoadConfig()
	if err != nil {
		logger.Fatalf("failed to initialize cfg: %s", err)
	}

	log := cfg.configureLogger()
	w := waiter.New(ctx)

	var certsStorage internal.CertificatesStorage
	mongoClient := cfg.initializeMongoClient(ctx)

	certsStorage, err = mongo.NewCertsStorage(ctx, mongodb.CollectionConfig{
		DB: mongoClient.Database("immune"),
	})
	if err != nil {
		logger.Fatalf("failed to create certs storage: %s", err)
	}

	var botDataStorage telegram.DataStorage = nil
	var botStateManager telegram.StateManager = nil
	redis.SetPrefix(Name)

	cache := redis.NewCache(cfg.initializeRedisClient(ctx))

	botDataStorage = botRedis.NewRedisDataStorage(cache)
	botStateManager = botRedis.NewStateStorage(cache)

	vaxService := vaxcert.NewService(certsStorage, cache)
	qrGen := bot.NewQRGenerator(cfg.QR.URLPattern)

	// Create telegram bot
	tgBotOts := []telegram.Option{
		telegram.WithCertificatesService(vaxService),
		telegram.WithQRGenerator(qrGen),
		telegram.WithDataStorage(botDataStorage),
		telegram.WithStateManager(botStateManager),
	}
	if cfg.Telegram.WebhookURL != "" && cfg.Telegram.ListenAddr != "" {
		// providing webhook URL and webhook addr means, that bot should listen to webhook updates
		// instead of long polling.
		tgBotOts = append(tgBotOts, telegram.WithWebhook(cfg.Telegram.WebhookURL, cfg.Telegram.ListenAddr))
	}
	tgBot, err := telegram.NewBot(cfg.Telegram.Token, tgBotOts...)

	if err != nil {
		logger.Fatalf("failed to run telegram bot: %s", err)
	}

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
		server.WithAddr(cfg.AppHTTPServer.Addr),
		server.WithTimeouts(serverTimeouts(cfg.AppHTTPServer.HTTPServerConfig)),
	)

	statusServer, ready := createStatusServer(cfg.StatusHTTPServer)

	w.Add(tgBot, appServer, statusServer)

	ready()
	if err := w.Run(); err != nil {
		log.Fatalf("error running server: %s", err)
	}

	log.Info("exiting")
}

func serverTimeouts(cfg HTTPServerConfig) server.Timeouts {
	return server.Timeouts{
		Write:    cfg.WriteTimeout,
		Read:     cfg.ReadTimeout,
		Idle:     cfg.IdleTimeout,
		Shutdown: cfg.ShutdownTimeout,
	}
}

func createStatusServer(cfg *StatusHTTPConfig) (*server.Server, func()) {
	readyCh := make(chan struct{}, 1)
	ready := func() {
		readyCh <- struct{}{}
	}

	return server.New(
		server.WithName("status"),
		server.WithAddr(cfg.Addr),
		server.WithHandler(status.NewHandler(readyCh)),
		server.WithTimeouts(serverTimeouts(cfg.HTTPServerConfig)),
	), ready
}
