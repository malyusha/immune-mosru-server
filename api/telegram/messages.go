package telegram

import (
	"fmt"
	"strings"

	"github.com/malyusha/immune-mosru-server/internal"
)

var (
	messageYourInvitesSuccess = "Вот коды, которые ты можешь дать своим друзьям, чтобы они тоже могли взаимодествовать со мной\n"
	messageActivated          = "Успех\n"
	messageNoSuchInvite       = "Что-то не сходится, ты точно уверен, что ввел код правильно?"
)

func createInvitesMessage(invites []internal.Invite) string {
	chunkSize := 5
	var b strings.Builder
	for ix, ivt := range invites {
		b.WriteString(fmt.Sprintf("*%s* ", ivt.Code))
		// write whitespace for chunk, otherwise space
		if (ix+1)%chunkSize == 0 {
			b.WriteString("\n")
		} else {
			b.WriteString(" ")
		}
	}

	return b.String()
}
