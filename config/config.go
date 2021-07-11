package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var env map[string]string

func init() {
	en, err := godotenv.Read(filepath.Join(os.Getenv("GOVENTY_ROOT"), "config/.env"))
	if err != nil {
		fmt.Print(err)
	}
	env = en
}

func ENV() map[string]string {
	return env
}
