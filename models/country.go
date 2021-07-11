package models

type Country struct {
	SharedModel
	Name string `json:"name" gorm:"not null;index:country_name"`
	IsoCode string `json:"iso_code" gorm:"not null;index:country_iso_code"`
	Users []User `json:"users"`
}