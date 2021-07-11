package models

type Tag struct {
	SharedModel
	Name string `json:"name" gorm:"not null;index:tag_name"`
	Description string `json:"description" gorm:"null"`
	/**
		* Has-Many relationships
	*/
	Taggables []Taggable `json:"-"`
}