package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/internal/users"
)

// Context represents a context of the current event. It stores data
// depending on its type, whether it's a message, callback or whatever.
type Context interface {
	context.Context
	// Message returns stored message if such presented.
	Message() *tb.Message

	// Callback returns stored callback if such presented.
	Callback() *tb.Callback

	// Query returns stored query if such presented.
	Query() *tb.Query

	// ChosenInlineResult returns stored inline result if such presented.
	ChosenInlineResult() *tb.ChosenInlineResult

	// ShippingQuery returns stored shipping query if such presented.
	ShippingQuery() *tb.ShippingQuery

	// PreCheckoutQuery returns stored pre checkout query if such presented.
	PreCheckoutQuery() *tb.PreCheckoutQuery

	// Poll returns stored poll if such presented.
	Poll() *tb.Poll

	// PollAnswer returns stored poll answer if such presented.
	PollAnswer() *tb.PollAnswer

	// Migration returns both migration from and to chat IDs.
	Migration() (int64, int64)

	// Sender returns the current recipient, depending on the context type.
	// Returns nil if user is not presented.
	Sender() *tb.User

	// Chat returns the current chat, depending on the context type.
	// Returns nil if chat is not presented.
	Chat() *tb.Chat

	// Recipient combines both Sender and Chat functions. If there is no user
	// the chat will be returned. The native context cannot be without sender,
	// but it is useful in the case when the context created intentionally
	// by the NewContext constructor and have only Chat field inside.
	Recipient() tb.Recipient

	// Text returns the message text, depending on the context type.
	// In the case when no related data presented, returns an empty string.
	Text() string

	// Data returns the current data, depending on the context type.
	// If the context contains command, returns its arguments string.
	// If the context contains payment, returns its payload.
	// In the case when no related data presented, returns an empty string.
	Data() string

	// Args returns a raw slice of command or callback arguments as strings.
	// The message arguments split by space, while the callback's ones by a "|" symbol.
	Args() []string

	// Send sends a message to the current recipient.
	// See Send from bot.go.
	Send(what interface{}, opts ...interface{}) error

	// SendAlbum sends an album to the current recipient.
	// See SendAlbum from bot.go.
	SendAlbum(a tb.Album, opts ...interface{}) error

	// Reply replies to the current message.
	// See Reply from bot.go.
	Reply(what interface{}, opts ...interface{}) error

	// Forward forwards the given message to the current recipient.
	// See Forward from bot.go.
	Forward(msg tb.Editable, opts ...interface{}) error

	// ForwardTo forwards the current message to the given recipient.
	// See Forward from bot.go
	ForwardTo(to tb.Recipient, opts ...interface{}) error

	// Edit edits the current message.
	// See Edit from bot.go.
	Edit(what interface{}, opts ...interface{}) error

	// EditCaption edits the caption of the current message.
	// See EditCaption from bot.go.
	EditCaption(caption string, opts ...interface{}) error

	// EditOrSend edits the current message if the update is callback,
	// otherwise the content is sent to the chat as a separate message.
	EditOrSend(what interface{}, opts ...interface{}) error

	// EditOrReply edits the current message if the update is callback,
	// otherwise the content is replied as a separate message.
	EditOrReply(what interface{}, opts ...interface{}) error

	// Delete removes the current message.
	// See Delete from bot.go.
	Delete() error

	// Notify updates the chat action for the current recipient.
	// See Notify from bot.go.
	Notify(action tb.ChatAction) error

	// Ship replies to the current shipping query.
	// See Ship from bot.go.
	Ship(what ...interface{}) error

	// Accept finalizes the current deal.
	// See Accept from bot.go.
	Accept(errorMessage ...string) error

	// Answer sends a response to the current inline query.
	// See Answer from bot.go.
	Answer(resp *tb.QueryResponse) error

	// Respond sends a response for the current callback query.
	// See Respond from bot.go.
	Respond(resp ...*tb.CallbackResponse) error

	// Get retrieves data from the context.
	Get(key string) interface{}

	// Set saves data in the context.
	Set(key string, val interface{})
}

// nativeContext is a native implementation of the Context interface.
// "context" is taken by context package, maybe there is a better name.
type nativeContext struct {
	b *Bot
	context.Context

	message            *tb.Message
	callback           *tb.Callback
	query              *tb.Query
	chosenInlineResult *tb.ChosenInlineResult
	shippingQuery      *tb.ShippingQuery
	preCheckoutQuery   *tb.PreCheckoutQuery
	poll               *tb.Poll
	pollAnswer         *tb.PollAnswer

	lock  sync.RWMutex
	store map[string]interface{}
}

// NewContextFromUpdate returns a new native context object,
// field by the passed update.
func (b *Bot) NewContextFromUpdate(upd tb.Update) (Context, error) {
	ctx := &nativeContext{
		Context:            context.Background(),
		b:                  b,
		message:            upd.Message,
		callback:           upd.Callback,
		query:              upd.Query,
		chosenInlineResult: upd.ChosenInlineResult,
		shippingQuery:      upd.ShippingQuery,
		preCheckoutQuery:   upd.PreCheckoutQuery,
		poll:               upd.Poll,
		pollAnswer:         upd.PollAnswer,
	}

	if ctx.Sender() == nil {
		// probably it's "Stop bot" action
		return nil, errors.New("failed to get sender of message")
	}
	sender := ctx.Sender()
	filter := users.SingleUserFilter{TelegramID: sender.ID}
	createData := users.CreateUserData{
		Username:  sender.Username,
		FirstName: sender.FirstName,
		LastName:  sender.LastName,
	}

	// Set default user data/state/chatstate to context.
	user, err := b.users.FindOrCreateUser(ctx, filter, createData)
	if err != nil {
		return ctx, fmt.Errorf("failed to find or create user: %w", err)
	}
	ctx = WithUser(ctx, user).(*nativeContext)

	state, err := b.getUserChatState(user.ID)
	if err != nil {
		return ctx, err
	}

	ctx = ContextWithState(ctx, state).(*nativeContext)

	chatData, err := b.getUserData(user.ID)
	if err != nil {
		return ctx, err
	}

	ctx = ContextWithUserData(ctx, chatData).(*nativeContext)
	ctx = SetBot(ctx, b).(*nativeContext)

	return ctx, nil
}

func (b *Bot) NewContextFromMessage(m *tb.Message) (Context, error) {
	return b.NewContextFromUpdate(tb.Update{Message: m})
}

func (b *Bot) NewContextFromCallback(cb *tb.Callback) (Context, error) {
	return b.NewContextFromUpdate(tb.Update{Callback: cb})
}

func (c *nativeContext) Message() *tb.Message {
	switch {
	case c.message != nil:
		return c.message
	case c.callback != nil:
		return c.callback.Message
	default:
		return nil
	}
}

func (c *nativeContext) Callback() *tb.Callback {
	return c.callback
}

func (c *nativeContext) Query() *tb.Query {
	return c.query
}

func (c *nativeContext) ChosenInlineResult() *tb.ChosenInlineResult {
	return c.chosenInlineResult
}

func (c *nativeContext) ShippingQuery() *tb.ShippingQuery {
	return c.shippingQuery
}

func (c *nativeContext) PreCheckoutQuery() *tb.PreCheckoutQuery {
	return c.preCheckoutQuery
}

func (c *nativeContext) Poll() *tb.Poll {
	return c.poll
}

func (c *nativeContext) PollAnswer() *tb.PollAnswer {
	return c.pollAnswer
}

func (c *nativeContext) Migration() (int64, int64) {
	return c.message.MigrateFrom, c.message.MigrateTo
}

func (c *nativeContext) Sender() *tb.User {
	switch {
	case c.message != nil:
		return c.message.Sender
	case c.callback != nil:
		return c.callback.Sender
	case c.query != nil:
		return &c.query.From
	case c.chosenInlineResult != nil:
		return &c.chosenInlineResult.From
	case c.shippingQuery != nil:
		return c.shippingQuery.Sender
	case c.preCheckoutQuery != nil:
		return c.preCheckoutQuery.Sender
	case c.pollAnswer != nil:
		return &c.pollAnswer.User
	default:
		return nil
	}
}

func (c *nativeContext) Chat() *tb.Chat {
	switch {
	case c.message != nil:
		return c.message.Chat
	case c.callback != nil:
		return c.callback.Message.Chat
	default:
		return nil
	}
}

func (c *nativeContext) Recipient() tb.Recipient {
	chat := c.Chat()
	if chat != nil {
		return chat
	}
	return c.Sender()
}

func (c *nativeContext) Text() string {
	switch {
	case c.message != nil:
		return c.message.Text
	case c.callback != nil:
		return c.callback.Message.Text
	default:
		return ""
	}
}

func (c *nativeContext) Data() string {
	switch {
	case c.message != nil:
		return c.message.Payload
	case c.callback != nil:
		return c.callback.Data
	case c.query != nil:
		return c.query.Text
	case c.chosenInlineResult != nil:
		return c.chosenInlineResult.Query
	case c.shippingQuery != nil:
		return c.shippingQuery.Payload
	case c.preCheckoutQuery != nil:
		return c.preCheckoutQuery.Payload
	default:
		return ""
	}
}

func (c *nativeContext) Args() []string {
	if c.message != nil {
		payload := strings.Trim(c.message.Payload, " ")
		if payload == "" {
			return nil
		}
		return strings.Split(payload, " ")
	}
	if c.callback != nil {
		return strings.Split(c.callback.Data, "|")
	}
	return nil
}

func (c *nativeContext) Send(what interface{}, opts ...interface{}) error {
	_, err := c.b.api.Send(c.Recipient(), what, opts...)
	return err
}

func (c *nativeContext) SendAlbum(a tb.Album, opts ...interface{}) error {
	_, err := c.b.api.SendAlbum(c.Recipient(), a, opts...)
	return err
}

func (c *nativeContext) Reply(what interface{}, opts ...interface{}) error {
	msg := c.Message()
	if msg == nil {
		return ErrBadContext
	}
	_, err := c.b.api.Reply(msg, what, opts...)
	return err
}

func (c *nativeContext) Forward(msg tb.Editable, opts ...interface{}) error {
	_, err := c.b.api.Forward(c.Recipient(), msg, opts...)
	return err
}

func (c *nativeContext) ForwardTo(to tb.Recipient, opts ...interface{}) error {
	msg := c.Message()
	if msg == nil {
		return ErrBadContext
	}
	_, err := c.b.api.Forward(to, msg, opts...)
	return err
}

func (c *nativeContext) Edit(what interface{}, opts ...interface{}) error {
	clb := c.callback
	if clb == nil || clb.Message == nil {
		return ErrBadContext
	}
	_, err := c.b.api.Edit(clb.Message, what, opts...)
	return err
}

func (c *nativeContext) EditCaption(caption string, opts ...interface{}) error {
	clb := c.callback
	if clb == nil || clb.Message == nil {
		return ErrBadContext
	}
	_, err := c.b.api.EditCaption(clb.Message, caption, opts...)
	return err
}

func (c *nativeContext) EditOrSend(what interface{}, opts ...interface{}) error {
	if c.callback != nil {
		return c.Edit(what, opts...)
	}
	return c.Send(what, opts...)
}

func (c *nativeContext) EditOrReply(what interface{}, opts ...interface{}) error {
	if c.callback != nil {
		return c.Edit(what, opts...)
	}
	return c.Reply(what, opts...)
}

func (c *nativeContext) Delete() error {
	msg := c.Message()
	if msg == nil {
		return ErrBadContext
	}
	return c.b.api.Delete(msg)
}

func (c *nativeContext) Notify(action tb.ChatAction) error {
	return c.b.api.Notify(c.Recipient(), action)
}

func (c *nativeContext) Ship(what ...interface{}) error {
	if c.shippingQuery == nil {
		return errors.New("telebot: context shipping query is nil")
	}
	return c.b.api.Ship(c.shippingQuery, what...)
}

func (c *nativeContext) Accept(errorMessage ...string) error {
	if c.preCheckoutQuery == nil {
		return errors.New("telebot: context pre checkout query is nil")
	}
	return c.b.api.Accept(c.preCheckoutQuery, errorMessage...)
}

func (c *nativeContext) Answer(resp *tb.QueryResponse) error {
	if c.query == nil {
		return errors.New("telebot: context inline query is nil")
	}
	return c.b.api.Answer(c.query, resp)
}

func (c *nativeContext) Respond(resp ...*tb.CallbackResponse) error {
	if c.callback == nil {
		return errors.New("telebot: context callback is nil")
	}
	return c.b.api.Respond(c.callback, resp...)
}

func (c *nativeContext) Set(key string, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.store == nil {
		c.store = make(map[string]interface{})
	}
	c.store[key] = value
}

func (c *nativeContext) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.store[key]
}

func SetBot(ctx Context, b *Bot) Context {
	ctx.Set("bot", b)
	return ctx
}

func GetBot(ctx Context) *Bot {
	return ctx.Get("bot").(*Bot)
}

func ContextWithState(ctx Context, state ChatState) Context {
	ctx.Set("chat-state", state)

	return ctx
}

func GetUserData(ctx Context) *UserData {
	data := ctx.Get("data")
	if c, ok := data.(*UserData); ok {
		return c
	}

	return new(UserData)
}

func GetState(ctx Context) ChatState {
	st := ctx.Get("chat-state")
	if st == nil {
		return ""
	}

	return st.(ChatState)
}

func WithUser(ctx Context, user *internal.User) Context {
	ctx.Set("user", user)

	return ctx
}

func GetUser(ctx Context) *internal.User {
	entry := ctx.Get("user")
	if user, ok := entry.(*internal.User); ok {
		return user
	}

	return nil
}

func ContextWithUserData(ctx Context, data *UserData) Context {
	ctx.Set("data", data)
	return ctx
}

var ErrBadContext = errors.New("bad context")
