package controllers

import (
	"bytes"
	_ "encoding/json"
	"fmt"
	"goventy/config"
	"goventy/models"
	"goventy/utils"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Home(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	events := make([]models.Event, 0)
	user.Find(1)
	models.DB().Model(&user).Related(&events)
	user.Events = events
	/*b, err := json.Marshal(user)
	if err != nil {
		fmt.Print(err)
	}*/
	parentPath := "/layouts/general.html"
	emailTemplate, err := template.New("layout").ParseFiles(
		filepath.Join(
			os.Getenv("goventy_ROOT"),
			filepath.Join(
				config.ENV()["MAIL_VIEWS_PATH"],
				parentPath)))
	emailTemplate.New("body").ParseFiles(filepath.Join(os.Getenv("goventy_ROOT"), config.ENV()["MAIL_VIEWS_PATH"]) + "/user/welcome.html")
	if err != nil {
		utils.Log(err, "error")
	}
	emailOutput := &bytes.Buffer{}
	w.Header().Set("Content-Type", "text/html")
	type TemplateData struct {
		Name     string
		Token    string
		FrontUrl string
		Year     int
	}
	td := TemplateData{
		Name:     strings.Title(strings.ToLower(user.Name)),
		Token:    user.EmailVerificationToken,
		FrontUrl: config.ENV()["FRONTEND_URL"],
		Year:     time.Now().UTC().Year(),
	}
	emailTemplate.ExecuteTemplate(emailOutput, "layout", td)
	fmt.Fprintf(w, "%s", emailOutput.String())
}
