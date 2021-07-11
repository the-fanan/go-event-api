package main

import (
	"fmt"
	"goventy/config"
	"goventy/models"
	"goventy/routes"
	"goventy/utils"
	"net/http"
)

func main() {
	router := routes.MakeRouter()
	defer models.DB().Close()
	fmt.Print("Starting server...")
	err := http.ListenAndServe(":"+config.ENV()["PORT"], router)
	if err != nil {
		utils.Log(err, "error")
	}
}
