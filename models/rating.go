package models

type Rating struct {
	SharedModel
	Like       uint   `json:"like" gorm:"null;type:boolean;default:0"`
	Dislike    uint   `json:"dislike" gorm:"null;type:boolean;default:0"`
	RatingType string `json:"rating_type" gorm:"not null;index:rating_type"`  //what are you  rating. user, place, goventy etc
	RatingID   uint   `json:"rating_id" gorm:"not null;index:rating_type_id"` //ID of the object you are rating
	UserID     uint   `json:"user_id" gorm:"not null"`
}
