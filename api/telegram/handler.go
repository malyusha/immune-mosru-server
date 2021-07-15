package telegram

import (
	"context"
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) handle(ctx Context) error {
	user := GetUser(ctx)
	state := GetState(ctx)
	if state == "" || state == Started {
		return nil
	}

	prevState := state
	handler := b.machine.GetHandlerForState(state)

	nextCtx, err := handler.Handle(ctx)
	if err != nil {
		return err
	}

	if !b.machine.isAllowedToTransit(prevState, GetState(nextCtx)) {
		return newInternalError(fmt.Errorf("transition from %s to %s not allowed", prevState, GetState(nextCtx)))
	}

	if GetState(nextCtx) != "" {
		err := b.dataStorage.SetData(ctx, user.ID, GetUserData(nextCtx))
		if err != nil {
			return newInternalError(fmt.Errorf("failed to set new data for handler: %w", err))
		}
	}

	if handler.OnExit != nil {
		if err := handler.OnExit(ctx); err != nil {
			handleError(ctx, newInternalError(fmt.Errorf("failed to reply: %w", err)))
		}
	}

	nextHandler, err := b.machine.TransitFrom(user.ID, prevState, GetState(nextCtx))
	if err != nil {
		return err
	}

	if nextHandler != nil && nextHandler.OnEnter != nil {
		if err := nextHandler.OnEnter(nextCtx); err != nil {
			b.machine.SetState(context.Background(), user.ID, GetState(nextCtx))
			b.dataStorage.SetData(context.Background(), user.ID, GetUserData(nextCtx))
			return fmt.Errorf("OnEnter call fail: %w", err)
		}
	}

	return nil
}

func (b *Bot) onText(m *tb.Message) {
	ctx, err := b.NewContextFromMessage(m)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleError(ctx, b.handle(ctx))
}

func (b *Bot) registerHandlers() {
	b.api.Handle(&btnMyInvites, func(m *tb.Message) {
		ctx, err := b.NewContextFromMessage(m)
		if err != nil {
			handleError(ctx, err)
		}

		user := GetUser(ctx)
		invites, err := b.users.GetUserInvites(ctx, user.ID)
		if err != nil {
			handleError(ctx, newInternalError(err))
			return
		}

		ctx.Send(createInvitesMessage(invites), tb.ModeMarkdownV2)
	})

	b.api.Handle(&btnAddMoreInvites, func(m *tb.Message) {
		ctx, err := b.NewContextFromMessage(m)
		if err != nil {
			handleError(ctx, err)
		}

		user := GetUser(ctx)
		if user.TelegramData.ID != b.adminUserId {
			return
		}

		user, err = b.users.AddInvitesToUser(ctx, user)
		if err != nil {
			handleError(ctx, newInternalError(err))
			return
		}

		ctx.Send(createInvitesMessage(user.Invites), tb.ModeMarkdownV2)
	})

	b.api.Handle(&btnActivateInvite, func(m *tb.Message) {
		ctx, err := b.NewContextFromMessage(m)
		if err != nil {
			handleError(ctx, err)
		}

		ctx = ContextWithState(ctx, InviteInit)
		handleError(ctx, b.handle(ctx))
	})
	b.api.Handle(&btnReceiveCode, func(m *tb.Message) {
		ctx, err := b.NewContextFromMessage(m)
		if err != nil {
			handleError(ctx, err)
		}

		ctx = ContextWithState(ctx, GenerateStart)

		handleError(ctx, b.handle(ctx))
	})

	// handle information button
	b.api.Handle(&btnInfo, func(m *tb.Message) {
		ctx, err := b.NewContextFromMessage(m)
		if err != nil {
			handleError(ctx, err)
			return
		}

		user := GetUser(ctx)

		md := fmt.Sprintf(`
*Информация*

Количество генераций осталось: %d
`, user.QRGenerationsLeft)
		handleError(ctx, ctx.Send(md, tb.ModeMarkdownV2))
	})

	b.api.Handle(tb.OnCallback, func(c *tb.Callback) {
		ctx, err := b.NewContextFromCallback(c)
		if err != nil {
			handleError(ctx, err)
			return
		}

		handleError(ctx, b.handle(ctx))
	})

	// Global text handler. Wraps default callback for message with sender state provision.
	b.api.Handle(tb.OnText, b.onText)
}
