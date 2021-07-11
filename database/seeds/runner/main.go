package main

import (
	"errors"
	"fmt"
	"goventy/database/seeds"
	"goventy/models"
	"os"
	"sort"

	"github.com/jinzhu/gorm"
)

func main() {
	type SeederFunction func()

	seeders := map[string]SeederFunction{
		"seeder_2021_01_18_1_tags":  seeds.Seeder_2021_01_18_1_tags,
		"seeder_2021_01_18_2_users": seeds.Seeder_2021_01_18_2_users,
	}

	//DO NOT ALTER CODE BEYOND THIS POINT
	if len(os.Args) <= 1 {
		//seed all tabled listed
		testSeeder := &models.Seeder{}
		count := 0
		err := models.DB().First(testSeeder).Count(&count).Error

		if err != nil {
			fmt.Println(err)
		}

		if count == 0 {
			//create migration and seeder table
			models.DB().Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
				&models.Seeder{},
			)
		}

		keys := make([]string, 0)
		for key, _ := range seeders {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			seeder := &models.Seeder{}
			//check if key exists in migration table
			err = models.DB().Where("seeder = ?", key).First(seeder).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				seeders[key]()
				seeder.Seeder = key
				err = models.DB().Create(seeder).Error
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	} else {
		//seed table speified
		//for this, run "go run main.go tags"
		arguments := os.Args
		seeders[arguments[1]]()
	}
}
