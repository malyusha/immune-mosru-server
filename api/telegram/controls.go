package telegram

import (
	"sync"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	menu = &tb.ReplyMarkup{ResizeReplyKeyboard: true}

	// OnExit buttons.
	btnInfo        = menu.Text("Информация")
	btnReceiveCode = menu.Text("Получить код")
)

func getMenu() *tb.ReplyMarkup {
	var once sync.Once

	once.Do(func() {
		menu.Reply(menu.Row(btnInfo, btnReceiveCode))
	})

	return menu
}
