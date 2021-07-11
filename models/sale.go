package models

type Sale struct {
	SharedModel
	OwnerType string `json:"owner_type" gorm:"not null;index:sale_owner"` //was it a ticket or a an item in quick items
	OwnerID uint `json:"owner_id" gorm:"not null;index:sale_owner_id"`
	UserID uint `json:"user_id" gorm:"null"`//user that paid
	Code string `json:"code" gorm:"not null"` //unique code for identifying this sale
	UserEmail string `json:"user_email" gorm:"null"`//if user is not logged in; user that paid
	UserName string `json:"user_name" gorm:"null"`
	/**
		* Belongs relationships
	*/
	Event Event `json:"event"`
	Ticket Ticket `json:"ticket"`
}