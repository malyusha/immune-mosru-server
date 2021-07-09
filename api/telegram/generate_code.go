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
)

const (
	Started                 ChatState = "started"
	GenerateStart           ChatState = "generate_start"
	CredentialsRequested    ChatState = "credentials_requested"
	DateBirthRequested      ChatState = "date_birth_requested"
	CredentialsConfirmation ChatState = "credentials_confirmation"
)

var generateStartStep = &StateHandler{
	Handle: func(ctx *Context) (*Context, error) {
		return ContextWithState(ctx, CredentialsRequested), nil
	},
}

var birthdayStep = &StateHandler{
	Title: "Дата рождения",
	OnEnter: func(ctx *Context) error {
		_, err := ctx.Bot().api.Send(ctx.GetSender(), "Теперь мне нужна твоя дата рождения в формате DD.MM.YYYY (например 01.01.2000)")
		return err
	},
	Handle: func(ctx *Context) (*Context, error) {
		reg := regexp.MustCompile(`\d{2}\.\d{2}\.\d{4}`)

		msg := ctx.GetMessage()
		if msg.Text == "" {
			return ctx, errors.New("Ты не понял, нужна твоя дата рождения")
		}

		if !reg.MatchString(msg.Text) {
			return ctx, errors.New("Неправильный формат, я же выше прислал, в каком формате нужна дата рождения, не тупи...")
		}

		data := ctx.GetData()

		data.Credentials.DateBirth = msg.Text

		ctx = ContextWithState(
			ContextWithData(ctx, data),
			CredentialsConfirmation,
		)

		return ctx, nil
	},
}

var confirmationStep = &StateHandler{
	OnEnter: func(ctx *Context) error {
		msg := `
Отлично, давай проверим, что все данные введены верно

<b>ФИО</b>: %s
<b>Дата рождения</b>: %s
`
		creds := ctx.GetData().Credentials
		fio := fmt.Sprintf("%s %s %s", creds.LastName, creds.FirstName, creds.SecondName)

		repl := &tb.ReplyMarkup{
			ResizeReplyKeyboard: true,
			ReplyKeyboardRemove: true,
			InlineKeyboard: [][]tb.InlineButton{
				{{Data: "credentials_ok", Text: "✅"}, {Data: "credentials_not_ok", Text: "❌"}},
			},
		}
		_, err := ctx.Bot().api.Send(ctx.GetSender(), fmt.Sprintf(msg, fio, creds.DateBirth), repl, tb.ModeHTML)

		return err
	},
	Handle: func(ctx *Context) (*Context, error) {
		cb := ctx.callback
		if cb == nil {
			return nil, errors.New("Непонятно как ты до этого дошел")
		}

		// delete inline keyboard
		_, err := ctx.Bot().api.EditReplyMarkup(cb.Message, &tb.ReplyMarkup{})
		if err != nil {
			return nil, err
		}

		switch cb.Data {
		case "credentials_not_ok":

			err = ctx.Bot().api.Respond(cb, &tb.CallbackResponse{Text: "Ок, давай по-новой"})

			return ContextWithState(ctx, CredentialsRequested), err
		case "credentials_ok":
			data := ctx.GetData()
			cert, err := ctx.Bot().vaxcert.CreateVaxCert(context.Background(), vaxcert.NewCertificateData{
				LastName:    data.Credentials.LastName,
				FirstName:   data.Credentials.FirstName,
				SecondName:  data.Credentials.SecondName,
				DateOfBirth: data.Credentials.DateBirth,
			})

			if err != nil {
				return ctx, errors.New("Что-то не получилось")
			}

			qr, err := ctx.Bot().qrgen.GenerateQR(cert.Code)
			if err != nil {
				return ctx, errors.New("Не получилось сгенерировать код")
			}

			photo := &tb.Photo{
				File:    tb.FromReader(bytes.NewReader(qr)),
				Caption: "Вот твой QR код",
			}
			if _, err := ctx.Bot().api.Send(ctx.GetSender(), photo, getMenu()); err != nil {
				return ctx, errors.New("Что-то не получилось")
			}

			data.QRRetries -= 1
			data.Credentials = UserCredentials{}

			ctx = ContextWithState(ctx, Started)
			return ContextWithData(ctx, data), nil
		default:
		}

		return ctx, nil
	},
}

var credentialsStep = &StateHandler{
	Title: "ФИО",
	OnEnter: func(ctx *Context) error {
		_, err := ctx.Bot().api.Send(ctx.GetSender(), "Для начала нужны твои ФИО", &tb.ReplyMarkup{ReplyKeyboardRemove: true})
		return err
	},
	Handle: func(ctx *Context) (*Context, error) {
		msg := ctx.GetMessage()
		creds := strings.Split(msg.Text, " ")
		if len(creds) != 3 {
			return ctx, errors.New("Чет не то, введи ФИО через пробел в правильном формате")
		}

		data := ctx.GetData()
		data.Credentials = UserCredentials{
			LastName:   creds[0],
			FirstName:  creds[1],
			SecondName: creds[2],
		}

		ctx = ContextWithState(ctx, DateBirthRequested)
		ctx = ContextWithData(ctx, data)
		return ctx, nil
	},
}

func createGenerateCommandMachine(manager StateManager) *Machine {
	machine := NewMachine(manager)
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
