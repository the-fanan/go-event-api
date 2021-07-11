package models

type Event struct {
	SharedModel
	UserID      uint   `json:"user_id" gorm:"not null"`
	Name        string `json:"name" gorm:"not null;index:event_name"`
	Description string `json:"description" gorm:"type:text;not null;"`
	Address     string `json:"address" gorm:"null;index:event_address"`
	StartDate   string `json:"start_date" gorm:"type:DATE;not null"`
	EndDate     string `json:"end_date" gorm:"type:DATE;not null"`
	Times       string `json:"times" gorm:"type:json;not null"`
	IsAvailable int    `json:"is_available" gorm:"type:int;not null;default:0"`

	/**
	* Belongs-To relationships
	 */
	Creator User `json:"creator" gorm:"foreignKey:UserID"`
	/**
	* Has-Many relationships
	 */
	Presenters []Presenter `json:"presenters"`
	/**
	* Polymorphic relationships
	 */
	AbuseReports []AbuseReport `json:"abuse_reports" gorm:"polymorphic:Report;"`
	Tickets      []Ticket      `json:"tickets" gorm:"polymorphic:Owner;"`
	Images       []Image       `json:"images" gorm:"polymorphic:Owner;"`
	Ratings      []Rating      `json:"ratings" gorm:"polymorphic:Rating;"`
	Reviews      []Review      `json:"reviews" gorm:"polymorphic:Review;"`
	Taggables    []Taggable    `json:"-" gorm:"polymorphic:Owner;"` //tags associated with this event

	/**
	* Many To Many
	 */
	Users []User `json:"users" gorm:"many2many:user_events;"` //users that have attended this goventy

	/**
	* Non-db related data
	 */
	Tags []Tag `json:"tags" gorm:"-"`
}
