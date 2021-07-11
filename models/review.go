package models

type Review struct {
	SharedModel
	Comment    string `json:"comment" gorm:"not null"`
	ReviewType string `json:"review_type" gorm:"not null;index:review_type"`  //What is this review for? user, goventy, place etc.
	ReviewID   uint   `json:"review_id" gorm:"not null;index:review_type_id"` //ID of what review if for
	UserID     uint   `json:"user_id" gorm:"not null"`
}
