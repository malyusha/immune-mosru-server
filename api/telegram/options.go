package telegram

import (
	"github.com/malyusha/immune-mosru-server/internal/users"
	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
)

type Option func(b *Bot)

func WithWebhook(url string, listenAddr string) Option {
	return func(b *Bot) {
		b.webhookParams = &webhookParams{
			addr:      listenAddr,
			publicURL: url,
		}
	}
}

func WithUsersService(srvc users.Service) Option {
	return func(b *Bot) {
		b.users = srvc
	}
}

func WithAdminUserId(id int) Option {
	return func(b *Bot) {
		b.adminUserId = id
	}
}

func WithCertificatesService(srvc vaxcert.Service) Option {
	return func(b *Bot) {
		b.vaxcert = srvc
	}
}

func WithQRGenerator(gen QRGen) Option {
	return func(b *Bot) {
		b.qrgen = gen
	}
}

func WithStateManager(manager StateManager) Option {
	return func(b *Bot) {
		b.stateManager = manager
	}
}

func WithDataStorage(storage DataStorage) Option {
	return func(b *Bot) {
		b.dataStorage = storage
	}
}
