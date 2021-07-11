package migrations

import (
	"goventy/models"
)

func Migration_2021_01_18_01_initial_migration() {
	models.DB().Exec("SET FOREIGN_KEY_CHECKS = 0")
	models.DB().DropTableIfExists(
		"abuse_reports",
		"reviews",
		"ratings",
		"images",
		"taggables",
		"tags",
		"sales",
		"tickets",
		"events",
		"payments",
		"users",
		"countries")
	models.DB().Exec("SET FOREIGN_KEY_CHECKS = 1")

	models.DB().Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Presenter{},
		&models.Ticket{},
		&models.Sale{},
		&models.Payment{},
		&models.Image{},
		&models.AbuseReport{},
		&models.Rating{},
		&models.Review{},
		&models.Tag{},
		&models.Taggable{},
		&models.Country{},
	)

	/**
	* Foreign keys
	 */
	models.DB().Model(&models.AbuseReport{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	models.DB().Model(&models.Event{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	models.DB().Model(&models.Payment{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	models.DB().Model(&models.Rating{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	models.DB().Model(&models.Review{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	models.DB().Model(&models.Taggable{}).AddForeignKey("tag_id", "tags(id)", "RESTRICT", "RESTRICT")
	models.DB().Model(&models.Presenter{}).AddForeignKey("event_id", "events(id)", "RESTRICT", "RESTRICT")
	models.DB().Model(&models.Sale{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
}
