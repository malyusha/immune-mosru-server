package telegram

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	menu = tb.ReplyMarkup{ResizeReplyKeyboard: true}

	// OnExit buttons.
	btnActivateInvite = menu.Text("Активировать инвайт код")
	btnMyInvites      = menu.Text("Мои инвайт-коды")
	btnInfo           = menu.Text("Информация")
	btnReceiveCode    = menu.Text("Получить код")
	btnAddMoreInvites = menu.Text("Добавить инвайтов")
)

func getMenu(ctx Context) *tb.ReplyMarkup {
	mainMenu := &tb.ReplyMarkup{ResizeReplyKeyboard: true}

	user := GetUser(ctx)
	if user.IsActive {
		row := tb.Row{btnInfo}
		if len(user.Invites) != 0 {
			row = append(row, btnMyInvites)
		}
		if user.QRGenerationsLeft != 0 {
			row = append(row, btnReceiveCode)
		}

		if user.TelegramData.ID == GetBot(ctx).adminUserId {
			mainMenu.Reply(row, tb.Row{btnAddMoreInvites})
		} else {
			mainMenu.Reply(row)
		}
	} else {
		mainMenu.Reply(
			menu.Row(btnActivateInvite),
		)
	}

	return mainMenu
}
