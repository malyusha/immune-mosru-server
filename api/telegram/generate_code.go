package telegram

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

const (
	Started                 ChatState = "started"
	GenerateStart           ChatState = "generate_start"
	CredentialsRequested    ChatState = "credentials_requested"
	DateBirthRequested      ChatState = "date_birth_requested"
	CredentialsConfirmation ChatState = "credentials_confirmation"
)

var generateStartStep = &StateHandler{
	Handle: func(ctx Context) (Context, error) {
		user := GetUser(ctx)

		if user.QRGenerationsLeft == 0 {
			if err := ctx.Send("Прости, но количество генераций кода ограничено"); err != nil {
				handleError(ctx, err)
			}

			// back to started state
			return ContextWithState(ctx, Started), nil
		}
		// proceed transition to request credentials
		return ContextWithState(ctx, CredentialsRequested), nil
	},
}

var birthdayStep = &StateHandler{
	Title: "Дата рождения",
	OnEnter: func(ctx Context) error {
		return ctx.Send("Теперь мне нужна твоя дата рождения в формате DD.MM.YYYY (например 01.01.2000)")
	},
	Handle: func(ctx Context) (Context, error) {
		reg := regexp.MustCompile(`\d{2}\.\d{2}\.\d{4}`)

		msg := ctx.Message()
		if msg.Text == "" {
			return ctx, errors.New("Ты не понял, нужна твоя дата рождения")
		}

		if !reg.MatchString(msg.Text) {
			return ctx, errors.New("Неправильный формат, я же выше прислал, в каком формате нужна дата рождения, не тупи...")
		}

		data := GetUserData(ctx)
		data.Credentials.DateBirth = msg.Text

		ctx = ContextWithState(
			ContextWithUserData(ctx, data),
			CredentialsConfirmation,
		)

		return ctx, nil
	},
}

var confirmationStep = &StateHandler{
	OnEnter: func(ctx Context) error {
		msg := `
Отлично, давай проверим, что все данные введены верно

<b>ФИО</b>: %s
<b>Дата рождения</b>: %s
`
		data := GetUserData(ctx)
		creds := data.Credentials
		fio := fmt.Sprintf("%s %s %s", creds.LastName, creds.FirstName, creds.SecondName)

		repl := &tb.ReplyMarkup{
			ResizeReplyKeyboard: true,
			ReplyKeyboardRemove: true,
			OneTimeKeyboard:     true,
			InlineKeyboard: [][]tb.InlineButton{
				{{Data: "credentials_ok", Text: "✅"}, {Data: "credentials_not_ok", Text: "❌"}},
			},
		}

		return ctx.Send(fmt.Sprintf(msg, fio, creds.DateBirth), repl, tb.ModeHTML)
	},
	Handle: func(ctx Context) (Context, error) {
		cb := ctx.Callback()
		data := ""
		if cb == nil {
			data = ctx.Message().Text
		} else {
			data = cb.Data
		}

		// delete inline keyboard
		if err := ctx.Edit(&tb.ReplyMarkup{}); err != nil {
			logger.Errorf("failed to delete inline keyboard: %s", err)
		}

		switch data {
		case "credentials_not_ok", "нет", "Нет":
			err := ctx.Respond(&tb.CallbackResponse{Text: "Ок, давай по-новой"})

			return ContextWithState(ctx, CredentialsRequested), err
		case "credentials_ok", "да", "Да":
			data := GetUserData(ctx)
			bot := GetBot(ctx)
			cert, err := bot.vaxcert.CreateVaxCert(context.Background(), vaxcert.NewCertificateData{
				OwnerID:     GetUser(ctx).ID,
				LastName:    data.Credentials.LastName,
				FirstName:   data.Credentials.FirstName,
				SecondName:  data.Credentials.SecondName,
				DateOfBirth: data.Credentials.DateBirth,
			})

			if err != nil {
				return ctx, newInternalError(err)
			}

			qr, err := bot.qrgen.GenerateQR(cert.Code)
			if err != nil {
				return ctx, newInternalError(err)
			}

			photo := &tb.Photo{
				File:    tb.FromReader(bytes.NewReader(qr)),
				Caption: "Вот твой QR код",
			}
			if err := ctx.Send(photo, getMenu(ctx)); err != nil {
				return ctx, newInternalError(err)
			}

			ctx = ContextWithState(ctx, Started)
			data.Credentials = UserCredentials{}
			return ContextWithUserData(ctx, data), nil
		default:
			return nil, errors.New("Да или нет?")
		}
	},
}

var credentialsStep = &StateHandler{
	Title: "ФИО",
	OnEnter: func(ctx Context) error {
		return ctx.Send("Для начала нужны твои ФИО", &tb.ReplyMarkup{ReplyKeyboardRemove: true})
	},
	Handle: func(ctx Context) (Context, error) {
		msg := ctx.Message()
		creds := strings.Split(msg.Text, " ")
		if len(creds) != 3 {
			return ctx, errors.New("Чет не то, введи ФИО через пробел в правильном формате")
		}

		data := GetUserData(ctx)
		data.Credentials = UserCredentials{
			LastName:   creds[0],
			FirstName:  creds[1],
			SecondName: creds[2],
		}

		ctx = ContextWithState(ctx, DateBirthRequested)
		ctx = ContextWithUserData(ctx, data)
		return ctx, nil
	},
}
