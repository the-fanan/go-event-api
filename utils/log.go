package utils

import (
	"fmt"
	"goventy/config"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Log(message interface{}, logLevel string) {

	year, month, day := time.Now().Date()
	//logfolder/year/month
	folder := fmt.Sprintf("%s/%s/%d/%d", config.ENV()["LOG_PATH"], logLevel, year, int(month))
	folder = filepath.Join(os.Getenv("goventy_ROOT"), folder)
	err := os.MkdirAll(folder, 0776)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(folder, fmt.Sprintf("%d.log", day)), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0776)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger := log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger := log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger := log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	switch logLevel {
	case "info":
		InfoLogger.Println(message)
	case "warn":
		WarningLogger.Println(message)
	case "error":
		ErrorLogger.Println(message)
	}
}
