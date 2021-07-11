package emails

import (
	"goventy/config"
	"goventy/utils"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SmtpSend(mailer *Mailer) {
	m := gomail.NewMessage()
	if mailer.From != "" {
		m.SetHeader("From", m.FormatAddress(mailer.From, "goventy"))
	} else {
		m.SetHeader("From", m.FormatAddress("contact@goventy.ng", "goventy"))
	}

	var cc []string
	if mailer.Cc != nil {
		cc = mailer.Cc
	} else {
		cc = []string{}
	}

	m.SetHeaders(map[string][]string{
		"To": mailer.To,
		"Cc": cc,
	})
	m.SetHeader("Subject", mailer.Subject)
	m.SetBody("text/html", mailer.Message)
	port, err := strconv.Atoi(config.ENV()["MAIL_PORT"])
	if err != nil {
		utils.Log(err, "error")
		return
	}
	//m.write
	d := gomail.NewDialer(config.ENV()["MAIL_HOST"], port, config.ENV()["MAIL_USERNAME"], config.ENV()["MAIL_PASSWORD"])
	err = d.DialAndSend(m)
	if err != nil {
		utils.Log(err, "error")
		return
	}
}
