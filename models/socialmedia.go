package models

type SocialMedia struct {
	SharedModel
	UserID uint `json:"user_id" gorm:"not null"`
	Provider string `json:"provider" gorm:"not null;index:social_media_provider"`
	Username string `json:"username" gorm:"null;"`//twitter and IG have username, facebook does not have
	UserToken string `json:"-" gorm:"not null;"`
	AccessToken string `json:"-" gorm:"not null;"`
	Url string `json:"url" gorm:"null;"`
}