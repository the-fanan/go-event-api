package models

type Taggable struct {
	SharedModel
	TagID uint `json:"tag_id" gorm:"not null"`
	OwnerType string `json:"owner_type" gorm:"not null;index:tag_owner"`//whho owns the tag?
	OwnerID uint `json:"owner_id" gorm:"not null;index:tag_owner_id"`//what is the ID of the owner
	/**
		* Belongs To relationships
	*/
	Tag Tag `json:"tag"`
}