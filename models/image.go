package models

type Image struct {
	SharedModel
	UserID uint `json:"user_id" gorm:"not null"` //the user that uploaded this image and owns it
	Url string `json:"url" gorm:"not null"`
	OwnerType string `json:"owner_type" gorm:"not null;index:image_owner"`
	OwnerID uint `json:"owner_id" gorm:"not null;index:image_owner_id"`
	Provider string `json:"provider" gorm:"not null; default:'local';index:image_provider"`
	ProviderID string `json:"provider_id" gorm:"null;index:image_provder_id"`
	Priority int `json:"priority" gorm:"not null;default:0"`
	Dimension string `json:"dimension" gorm:"not null;default:'2D'"`
}