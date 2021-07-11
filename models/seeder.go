package models

type Seeder struct {
	ID uint `json:"id" gorm:"primary_key"`
	Seeder string `json:"seeder" gorm:"not null"`
}

