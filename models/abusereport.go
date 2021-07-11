package models

type AbuseReport struct {
	SharedModel
	Comment    string `json:"comment" gorm:"not null"`
	ReportType string `json:"report_type" gorm:"not null;index:report_type"` //report on a goventy, place, user, etc.
	ReportID   uint   `json:"report_id" gorm:"not null;index:report_id"`     //ID of item you are reporting
	UserID     uint   `json:"user_id"`
}
