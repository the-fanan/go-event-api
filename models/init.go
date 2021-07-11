package models

import (
	"fmt"
	"goventy/config"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func init() {
	uri := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", config.ENV()["DB_USERNAME"], config.ENV()["DB_PASSWORD"], config.ENV()["DB_HOSTNAME"], config.ENV()["DB_PORT"], config.ENV()["DB_DATABASE"])
	d, err := gorm.Open("mysql", uri)

	if err != nil {
		fmt.Print(err)
	}

	db = d
}

func DB() *gorm.DB {
	return db
}

type SharedModel struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
