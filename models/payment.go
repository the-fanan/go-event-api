package models

type Payment struct {
	SharedModel
	UserID uint `json:"user_id" gorm:"null"`//user that paid
	UserEmail string `json:"user_email" gorm:"null"`//if user is not logged in email of user that paid
	UserName string `json:"user_name" gorm:"null"`//if user is  not logged in name of user
	OwnerType string `json:"owner_type" gorm:"not null;index:payment_owner"` //Is this payment from a sale or subscription for stuff?
	OwnerID uint `json:"owner_id" gorm:"not null;index:payment_owner_id"`
	Amount float64 `json:"amount" gorm:"not null"`
	Details string `json:"details" gorm:"not null"`
	Provider string `json:"provider" gorm:"not null;"`//paystack, paypal, flutterwave, stripe etc.
	ProviderReference string `json:"provider_reference" gorm:"not null"`//reference from source
}