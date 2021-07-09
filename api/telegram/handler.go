package telegram

import (
	"context"
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) handleMessage(ctx *Context) error {
	senderID := ctx.GetSender().ID
	if ctx.GetMessage().Text == "/start" {
		b.stateManager.SetState(context.Background(), senderID, Started)
		_, err := b.api.Send(ctx.GetSender(), "Выбери что нужно", getMenu())
		return err
	}

	if ctx.GetState() == "" || ctx.GetState() == Started {
		if ans := b.chatter.answerFor(ctx.GetSender().ID); ans != "" {
			b.api.Send(ctx.GetSender(), ans)
		}

		return nil
	}

	prevState := ctx.GetState()
	handler := b.machine.GetHandlerForState(ctx.GetState())

	nextCtx, err := handler.Handle(ctx)
	if err != nil {
		_, err := b.api.Send(ctx.GetSender(), err.Error())
		return err
	}

	if !b.machine.isAllowedToTransit(prevState, nextCtx.GetState()) {
		return fmt.Errorf("transition from %s to %s not allowed", prevState, nextCtx.GetState())
	}

	if nextCtx.GetData() != nil {
		err := b.dataStorage.SetData(context.Background(), senderID, nextCtx.GetData())
		if err != nil {
			return fmt.Errorf("failed to set new data for handler: %w", err)
		}
	}

	if handler.OnExit != nil {
		if err := handler.OnExit(ctx); err != nil {
			b.handleError(ctx.GetMessage().Sender, fmt.Errorf("failed to reply: %w", err))
		}
	}

	nextHandler, err := b.machine.TransitFrom(senderID, prevState, nextCtx.GetState())
	if err != nil {
		return err
	}

	if nextHandler != nil && nextHandler.OnEnter != nil {
		if err := nextHandler.OnEnter(nextCtx); err != nil {
			b.machine.SetState(context.Background(), senderID, nextCtx.GetState())
			b.dataStorage.SetData(context.Background(), senderID, nextCtx.GetData())
			return fmt.Errorf("OnEnter call fail: %w", err)
		}
	}

	return nil
}

func (b *Bot) onText(m *tb.Message) {
	ctx, err := b.newContext(m, m.Sender)
	if err != nil {
		b.handleError(m.Sender, err)
		return
	}

	if err := b.handleMessage(ctx); err != nil {
		b.handleError(m.Sender, err)
	}
}

func (b *Bot) newContext(m *tb.Message, u *tb.User) (*Context, error) {
	currentState, err := b.getUserChatState(u)
	if err != nil {
		return nil, err
	}

	currentData, err := b.getUserData(u)
	if err != nil {
		return nil, err
	}

	return &Context{
		bot:     b,
		message: m,
		user:    u,
		state:   currentState,
		data:    currentData,
	}, nil
}

func (b *Bot) registerHandlers() {
	b.api.Handle(&btnReceiveCode, func(m *tb.Message) {
		ctx, err := b.newContext(m, m.Sender)
		if err != nil {
			b.handleError(m.Sender, err)
			return
		}

		if ctx.GetData().QRRetries == 0 {
			if _, err := b.api.Send(m.Sender, fmt.Sprintf("Сорян, но максимальное количество кодов - %d.", maxQRGenerations)); err != nil {
				b.handleError(m.Sender, err)
			}
			return
		}

		ctx = ContextWithState(ctx, GenerateStart)

		b.handleError(m.Sender, b.handleMessage(ctx))
	})

	b.api.Handle(&btnInfo, func(m *tb.Message) {
		data, err := b.getUserData(m.Sender)
		if err != nil {
			b.handleError(m.Sender, err)
			return
		}

		md := fmt.Sprintf(`
*Информация*

Количество генераций осталось: %d
`, data.QRRetries)
		if _, err = b.api.Send(m.Sender, md, tb.ModeMarkdownV2); err != nil {
			b.handleError(m.Sender, err)
			return
		}
	})

	b.api.Handle(tb.OnCallback, func(c *tb.Callback) {
		ctx, err := b.newContext(c.Message, c.Sender)
		if err != nil {
			b.handleError(c.Sender, err)
			return
		}
		ctx.callback = c
		b.handleError(c.Sender, b.handleMessage(ctx))
	})

	// Global text handler. Wraps default callback for message with user state provision.
	b.api.Handle(tb.OnText, b.onText)
}
