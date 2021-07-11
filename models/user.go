package models

type User struct {
	SharedModel
	Name                   string `json:"name" gorm:"not null;index:user_name"`
	Username               string `json:"username" gorm:"unique_index;not null;"`
	Email                  string `json:"email" gorm:"unique_index;not null;"`
	Password               string `json:"-" gorm:"not null"`
	EmailVerified          uint   `json:"email_verified" gorm:"type:boolean;default:0"`
	EmailVerificationToken string `json:"-" gorm:"NULL"`
	PasswordResetToken     string `json:"-" gorm:"NULL"`
	RememberToken          string `json:"remember_token" gorm:"NULL"`
	CountryID              uint   `json:"country_id" gorm:"null;index:country_id"`
	/**
	* Has-Many relationships
	 */
	Events         []Event       `json:"events"`
	Purchases      []Sale        `json:"purchases"`
	AbusesReported []AbuseReport `json:"abuses_reported"` //app abuse you report
	Payments       []Payment     `json:"payments"`
	Ratings        []Rating      `json:"ratings_given"` //ratings user has given other people
	Reviews        []Review      `json:"reviews_given"` //reviews user has given
	/**
	* Polymorphic relationships
	 */
	Images       []Image       `json:"images" gorm:"polymorphic:Owner;"`
	AbuseReports []AbuseReport `json:"abuse_reports" gorm:"polymorphic:Report;"` //abuse reports made against you
	/**
	* Many To Many
	 */
	AttendedEvents []Event `json:"attended_events" gorm:"many2many:user_events;"`
}

func (u *User) Find(id uint) {
	if id > 0 {
		db.First(&u, id)
	}
}
