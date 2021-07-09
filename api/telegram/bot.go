package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

const (
	maxQRGenerations = 5
)

type DataStorage interface {
	GetData(ctx context.Context, id int) (*UserData, error)
	SetData(ctx context.Context, id int, state *UserData) error
}

type QRGen interface {
	GenerateQR(data string) ([]byte, error)
}

type Bot struct {
	qrgen   QRGen
	vaxcert vaxcert.Service
	logger  logger.Logger
	api     *tb.Bot

	chatter *chatter

	webhookParams *webhookParams
	machine       *Machine
	stateManager  StateManager
	dataStorage   DataStorage
}

type webhookParams struct {
	addr      string
	publicURL string
}

type UserCredentials struct {
	FirstName  string
	LastName   string
	SecondName string
	DateBirth  string
}

func (b *Bot) Start(ctx context.Context) chan error {
	log := b.logger.WithContext(ctx)
	done := make(chan error)

	log.Infof("Running tg bot")
	go b.api.Start()
	go func() {
		<-ctx.Done()
		if _, ok := b.api.Poller.(*tb.Webhook); ok {
			if err := b.api.RemoveWebhook(); err != nil {
				b.logger.Errorf("failed to remove webhook: %s", err)
			}
		}
		b.api.Stop()
		done <- nil
	}()

	return done
}

func NewBot(token string, opts ...Option) (*Bot, error) {
	settings := tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tb.NewBot(settings)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		logger: logger.GetLogger(),
		api:    b,
	}

	for _, o := range opts {
		o(bot)
	}

	if bot.webhookParams != nil {
		bot.api.Poller = &tb.Webhook{
			Listen:         bot.webhookParams.addr,
			MaxConnections: 20,
			HasCustomCert:  false,
			PendingUpdates: 0,
			ErrorUnixtime:  0,
			ErrorMessage:   "",
			Endpoint: &tb.WebhookEndpoint{
				PublicURL: bot.webhookParams.publicURL,
			},
		}
	}

	if bot.dataStorage == nil {
		bot.logger.Warn("Running bot with in-memory data storage. Storage is not persisted. Ensure this is correct configuration.")
		bot.dataStorage = newInmemStorage()
	}

	if bot.stateManager == nil {
		bot.logger.Warn("Running bot with in memory state manager. StateHandler is not persisted. Ensure this is correct configuration.")
	}

	if err := validate(bot); err != nil {
		return nil, fmt.Errorf("invalid bot configuration: %w", err)
	}

	bot.machine = createGenerateCommandMachine(bot.stateManager)
	bot.registerHandlers()

	defaultResponses := []string{
		"Не понимаю что тебе нужно",
		"Вот тут внизу кнопки - это все что я могу сделать в данный момент",
		"Я не самый лучший собеседник",
		"Спроси себя - зачем ты тут, затем действуй",
		"Вспомни обо мне когда тебя не пустят попить {твой_любимый_хипстерский_коктейл} в пятницу вечером",
		"Вот у ребят уже есть код, а у тебя?",
	}

	lastResponses := []string{
		"Я устал, я ухожу",
		"Больше не хочу говорить с тобой без дела, считай это последнее предупреждение",
		"Так, разговор наш потихоньку заходит в тупик, давай по делу",
	}

	bot.chatter = newChatter(defaultResponses, lastResponses)
	return bot, nil
}

func (b *Bot) getUserChatState(u *tb.User) (ChatState, error) {
	ctx := context.Background()
	st, err := b.stateManager.GetState(ctx, u.ID)
	if err != nil {
		if err != ErrStateMissing {
			return "", fmt.Errorf("failed to gen chat state: %w", err)
		}

		st = Started
		if err = b.stateManager.SetState(ctx, u.ID, st); err != nil {
			return "", fmt.Errorf("failed to set initial state of chat for user ID %q: %w", u.ID, err)
		}
	}

	return st, nil
}

func (b *Bot) getUserData(u *tb.User) (*UserData, error) {
	ctx := context.Background()
	st, err := b.dataStorage.GetData(ctx, u.ID)
	if err != nil {
		if err != ErrDataMissing {
			return nil, err
		}

		// create new base state when state is missing
		st = InitialUserData()
		if err := b.dataStorage.SetData(ctx, u.ID, st); err != nil {
			return nil, fmt.Errorf("failed to create new base state: %w", err)
		}
	}

	return st, nil
}

func (b *Bot) handleError(recipient tb.Recipient, err error) {
	if err == nil {
		return
	}
	logger.Errorf("handle error: %s", err)

	if _, err := b.api.Send(recipient, "Что-то со мной не так. Попробуй еще разок"); err != nil {
		logger.Errorf("failed to send fail message to user: %s", err)
	}
}

// validates bot instance.
func validate(bot *Bot) error {
	if bot.stateManager == nil {
		return errors.New("no state manager provided for bot")
	}

	if bot.qrgen == nil {
		return errors.New("no GQ generation service provided")
	}

	if bot.vaxcert == nil {
		return errors.New("no certificate service provided")
	}

	return nil
}
