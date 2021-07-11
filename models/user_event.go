package models

type UserEvent struct {
	SharedModel
	UserID   uint `json:"user_id" gorm:"not null"`
	EventID  uint `json:"goventy_id" gorm:"not null"`
	RatingID uint `json:"rating_id" gorm:"null"`
	ReviewID uint `json:"review_id" gorm:"null"`

	/**
	* Polymorphic relationships
	 */
	Images []Image `json:"images" gorm:"polymorphic:Owner;"` //images the user uploaded on this event -- possible future feature
}
