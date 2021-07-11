package emails

import (
	"bytes"
	"errors"
	"goventy/config"
	"goventy/utils"
	"html/template"
	"os"
	"path/filepath"
)

type Mailer struct {
	TemplateData       interface{}
	TemplatePath       string
	TemplateParentPath string
	To                 []string
	From               string
	Cc                 []string
	Bcc                []string
	Subject            string
	Message            string
}

func (mailer *Mailer) GenerateTemplate() error {
	parentPath := "/layouts/general.html"
	if mailer.TemplateParentPath != "" {
		parentPath = mailer.TemplateParentPath
	}

	if mailer.TemplatePath == "" {
		return errors.New("No template path provided")
	}

	emailTemplate, err := template.New("layout").ParseFiles(
		filepath.Join(
			os.Getenv("goventy_ROOT"),
			filepath.Join(
				config.ENV()["MAIL_VIEWS_PATH"],
				parentPath)))
	emailTemplate.New("body").ParseFiles(
		filepath.Join(
			os.Getenv("goventy_ROOT"),
			filepath.Join(
				config.ENV()["MAIL_VIEWS_PATH"],
				mailer.TemplatePath)))
	if err != nil {
		return err
	}
	emailOutput := &bytes.Buffer{}
	emailTemplate.ExecuteTemplate(emailOutput, "layout", mailer.TemplateData)
	mailer.Message = emailOutput.String()
	return nil
}

func (mailer *Mailer) Send() {
	err := mailer.GenerateTemplate()
	if err != nil {
		utils.Log(err, "error")
		return
	}

	driver := config.ENV()["MAIL_DRIVER"]

	if driver == "log" {
		utils.Log(mailer.Message, "info")
		return
	}

	switch driver {
	case "smtp":
		SmtpSend(mailer)
		break
	case "mailgun":
		MailgunSend(mailer)
		break
	default:
		utils.Log("Driver"+driver+" is not supported", "error")
	}
}
