package emails

import (
	"context"
	"fmt"
	"goventy/config"
	"goventy/utils"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

func MailgunSend(mailer *Mailer) {
	mg := mailgun.NewMailgun(config.ENV()["MAILGUN_DOMAIN"], config.ENV()["MAILGUN_PRIVATE_KEY"])

	from := "goventy <contact@goventy.ng>"
	if mailer.From != "" {
		from = mailer.From
	}

	message := mg.NewMessage(from, mailer.Subject, "", mailer.To...)
	message.SetHtml(mailer.Message)
	for _, cc := range mailer.Cc {
		message.AddCC(cc)
	}

	for _, bcc := range mailer.Bcc {
		message.AddBCC(bcc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		utils.Log(err, "error")
	}

	utils.Log(fmt.Sprintf("ID: %s Resp: %s\n", id, resp), "info")
}
