package main

import (
	"errors"
	"fmt"
	"goventy/database/migrations"
	"goventy/models"
	"sort"

	"github.com/jinzhu/gorm"
)

func main() {
	type MigratorFunction func()

	migrators := map[string]MigratorFunction{
		//migration_YEAR_MONTH_DAY_NthMigrationForThatDay_NameOfMigration
		"migration_2021_01_18_1_initial_migration": migrations.Migration_2021_01_18_01_initial_migration,
	}

	//DO NOT ALTER CODE BEYOND THIS POINT
	testMigration := &models.Migration{}
	count := 0
	err := models.DB().First(testMigration).Count(&count).Error

	if err != nil {
		fmt.Println(err)
	}

	if count == 0 {
		//create migration and seeder table
		models.DB().Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
			&models.Migration{},
		)
	}

	keys := make([]string, 0)
	for key, _ := range migrators {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		migration := &models.Migration{}
		//check if key exists in migration table
		err = models.DB().Where("migration = ?", key).First(migration).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			migrators[key]()
			migration.Migration = key
			err = models.DB().Create(migration).Error
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
