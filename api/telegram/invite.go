package telegram

import (
	"errors"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/malyusha/immune-mosru-server/internal/users"
)

const (
	InviteInit      ChatState = "invite_init"
	InviteRequested ChatState = "invite_requested"
)

var inviteInitStep = &StateHandler{
	Handle: func(ctx Context) (Context, error) {
		user := GetUser(ctx)
		if user.IsActive {
			ctx.Send("Ты уже активировал инвайт, зачем тебе еще?")
			// user already activated invite. no need to request invite again
			return ContextWithState(ctx, Started), nil
		}
		return ContextWithState(ctx, InviteRequested), nil
	},
}

var inviteInputStep = &StateHandler{
	OnEnter: func(ctx Context) error {
		return ctx.Send("Введи инвайт-код, полученный непонятным образом", &tb.ReplyMarkup{ReplyKeyboardRemove: true})
	},

	Handle: func(ctx Context) (Context, error) {
		code := ctx.Message().Text
		bot := GetBot(ctx)
		currentUser := GetUser(ctx)

		user, err := bot.users.ActivateInvite(ctx, currentUser.ID, code)
		if err != nil {
			if errors.Is(err, users.ErrNoInvite) {
				return nil, errors.New(messageNoSuchInvite)
			}

			return ctx, newInternalError(err)
		}

		// write new user data to ctx
		ctx = WithUser(ctx, user)
		var msg = messageActivated
		if len(user.Invites) > 0 {
			msg += messageYourInvitesSuccess + createInvitesMessage(user.Invites)
		}
		
		if err := ctx.Send(msg, tb.ModeMarkdownV2, getMenu(ctx)); err != nil {
			return nil, newInternalError(err)
		}

		return ContextWithState(ctx, Started), nil
	},
}
