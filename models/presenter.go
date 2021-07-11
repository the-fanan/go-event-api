package models

/**
	* This model tells you who is going to be speaking or singing or enteratining at the event
	*/
type Presenter struct {
	SharedModel
	UserID uint `json:"user_id" gorm:"not null"` //the user that uploaded this
	EventID uint `json:"event_id" gorm:"not null"`
	Name string `json:"name" gorm:"not null;index:presenter_name"`
	Topic string `json:"topic" gorm:"null"`
	IsAvailable int  `json:"is_available" gorm:"type:int;not null;default:0"`
	/**
		* Polymorphic relationships
	*/
	Images   []Image `json:"images" gorm:"polymorphic:Owner;"`//image of presenter
	/**
		* Belongs relationships
	*/
	Event Event `json:"event" gorm:"foreignKey:EventID"`
	Creator User `json:"creator" gorm:"foreignKey:UserID"`
}