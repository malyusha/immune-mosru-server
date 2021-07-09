package telegram

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

type Context struct {
	bot *Bot

	user     *tb.User
	callback *tb.Callback
	message  *tb.Message
	state    ChatState
	data     *UserData
}

func NewContext(bot *Bot, message *tb.Message, user *tb.User) *Context {
	ctx := &Context{
		bot:     bot,
		user:    user,
		message: message,
	}

	return ctx
}

func (c *Context) Bot() *Bot {
	return c.bot
}

func (c *Context) GetMessage() *tb.Message {
	return c.message
}

func (c *Context) GetSender() *tb.User {
	return c.user
}

func (c *Context) GetState() ChatState {
	if c.state == "" {
		return Started
	}

	return c.state
}

func (c *Context) GetData() *UserData {
	if c.data == nil {
		return InitialUserData()
	}
	return c.data
}

func ContextWithData(ctx *Context, data *UserData) *Context {
	ctx.data = data
	return ctx
}

func ContextWithState(ctx *Context, state ChatState) *Context {
	ctx.state = state

	return ctx
}
