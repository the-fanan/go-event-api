package models

type Ticket struct {
	SharedModel
	UserID            uint    `json:"user_id" gorm:"not null"`                       //the user that uploaded this
	OwnerType         string  `json:"owner_type" gorm:"not null;index:ticket_owner"` //Is this ticket for an event or a Paid goventy or something else
	OwnerID           uint    `json:"owner_id" gorm:"not null;index:ticket_owner_id"`
	Name              string  `json:"name" gorm:"not null;default:'regular';index:ticket_name"` //[vip,regular etc.]
	Description       string  `json:"description" gorm:"NULL;type:text;"`
	Type              string  `json:"type" gorm:"not null;default:'purchased'"`      //[purchased, invite]
	Amount            float64 `json:"amount" gorm:"not null;default:0"`              //if amount is 0 then it is free
	Cost              float64 `json:"cost" gorm:"not null;default:0"`                //for exclusive events
	QuantityAvailable int     `json:"quantity_available" gorm:"not null; default:0"` //how many tickets should be available. 0 for infinity
	IsAvailable       int     `json:"is_available" gorm:"type:int;not null;default:0"`
	/**
	* Has-Many relationships
	 */
	Sales []Sale `json:"sales"`

	/**
	* Belongs relationships
	 */
	Creator User `json:"creator" gorm:"foreignKey:UserID"`

	/**
	* Polymorphic relationships
	 */
	Images []Image `json:"images" gorm:"polymorphic:Owner;"`
}
