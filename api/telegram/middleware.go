package telegram

import (
	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

func (b *Bot) NewAuthMiddleware() func(*tb.Update) bool {
	return func(update *tb.Update) bool {
		if update == nil {
			return false
		}
		ctx, err := b.NewContextFromUpdate(*update)
		if err != nil {
			handleError(ctx, err)
			return false
		}

		user := GetUser(ctx)
		if !user.IsActive {
			if GetState(ctx) == InviteRequested {
				return true
			}

			if ctx.Message().Text == btnActivateInvite.Text {
				b.stateManager.SetState(ctx, user.ID, InviteInit)
				return true
			}

			chatData := GetUserData(ctx)
			if chatData.InviteNotificationSent {
				ctx.Send("Начнем работать сразу после того, как отправишь инвайт-код", getMenu(ctx))
				// do not process updates because user has been notified, that invite code is must-have to work with bot.
				return false
			}

			if err := ctx.Send(msgNeedInviteCode, getMenu(ctx)); err != nil {
				handleError(ctx, err)
			} else {
				chatData.InviteNotificationSent = true
				if err := b.dataStorage.SetData(ctx, user.ID, chatData); err != nil {
					logger.
						With(logger.Fields{"err": err}).
						Error("failed to store new chat data after invite-code notification sent")
				}
			}

			return false
		}

		// user is activated, just sending start command again to reset state maybe
		if ctx.Message().Text == startCommand {
			b.stateManager.SetState(ctx, user.ID, Started)
			ctx.Send("Выбери что нужно", getMenu(ctx))
			return false
		}

		return true
	}
}

var (
	msgNeedInviteCode = "Привет! Для того чтобы начать, тебе нужно отправить инвайт-код. " +
		"Этот код тебе может дать твой друг/знакомый. К сожалению, без " +
		"такого кода я не работаю, так как не хочу чтобы мной пользовались без ограничений."
)
