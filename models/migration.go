package models

type Migration struct {
	ID uint `json:"id" gorm:"primary_key"`
	Migration string `json:"migration" gorm:"not null"`
}

