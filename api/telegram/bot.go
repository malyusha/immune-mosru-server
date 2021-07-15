package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/malyusha/immune-mosru-server/internal/users"
	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

const (
	startCommand  = "/start"
)

type DataStorage interface {
	GetData(ctx context.Context, id string) (*UserData, error)
	SetData(ctx context.Context, id string, state *UserData) error
}

type QRGen interface {
	GenerateQR(data string) ([]byte, error)
}

type Bot struct {
	adminUserId int
	qrgen   QRGen
	users   users.Service
	vaxcert vaxcert.Service
	logger  logger.Logger
	api     *tb.Bot

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
		/*Reporter: func(err error) {
			logger.Errorf("bot error: %s", err.Error())
		},*/
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

	bot.api.Poller = tb.NewMiddlewarePoller(bot.api.Poller, bot.NewAuthMiddleware())

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

	return bot, nil
}

func (b *Bot) getUserChatState(id string) (ChatState, error) {
	ctx := context.Background()
	st, err := b.stateManager.GetState(ctx, id)
	if err != nil {
		if err != ErrStateMissing {
			return "", fmt.Errorf("failed to gen chat state: %w", err)
		}

		st = Started
		if err = b.stateManager.SetState(ctx, id, st); err != nil {
			return "", fmt.Errorf("failed to set initial state of chat for user ID %q: %w", id, err)
		}
	}

	return st, nil
}

func (b *Bot) getUserData(id string) (*UserData, error) {
	ctx := context.Background()
	st, err := b.dataStorage.GetData(ctx, id)
	if err != nil {
		if err != ErrDataMissing {
			return nil, err
		}

		// create new base state when state is missing
		st = InitialUserData()
		if err := b.dataStorage.SetData(ctx, id, st); err != nil {
			return nil, fmt.Errorf("failed to create new base state: %w", err)
		}
	}

	return st, nil
}

// returns new state machine for manipulating commands and chat state.
func createGenerateCommandMachine(manager StateManager) *Machine {
	machine := NewMachine(manager)

	// enter invite command
	machine.AddTransitions(InviteInit, Started)
	machine.AddTransitions(InviteInit, InviteRequested)
	machine.AddTransitions(InviteRequested, Started)

	machine.AddStateHandler(InviteInit, inviteInitStep)
	machine.AddStateHandler(InviteRequested, inviteInputStep)

	// generate QR code command
	machine.AddTransitions(GenerateStart, Started) // because user can exceed number or qr code generations.
	machine.AddTransitions(GenerateStart, CredentialsRequested)
	machine.AddTransitions(CredentialsRequested, DateBirthRequested)
	machine.AddTransitions(DateBirthRequested, CredentialsConfirmation)
	machine.AddTransitions(CredentialsConfirmation, CredentialsRequested, Started)

	machine.AddStateHandler(GenerateStart, generateStartStep)
	machine.AddStateHandler(CredentialsRequested, credentialsStep)
	machine.AddStateHandler(DateBirthRequested, birthdayStep)
	machine.AddStateHandler(CredentialsConfirmation, confirmationStep)

	return machine
}

func handleError(ctx Context, err error) {
	if err == nil {
		return
	}

	log := logger.WithContext(ctx)

	var msg string
	if IsInternal(err) {
		log.Error(err.Error())
		msg = "Что-то со мной не так. Попробуй еще разок"
	} else {
		msg = err.Error()
	}

	if ctx != nil {
		if err := ctx.Send(msg); err != nil {
			logger.Errorf("failed to send fail message to sender: %s", err)
		}
	} else {
		logger.Errorf("context is nil. error received: %s", err)
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
