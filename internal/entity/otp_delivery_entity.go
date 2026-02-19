package entity

import "time"

// OtpDelivery tracks the delivery status of each OTP send attempt.
// Maps to table: otp_deliveries
type OtpDelivery struct {
	ID           string     `gorm:"column:id;primaryKey"`
	OtpID        string     `gorm:"column:otp_id"`
	Channel      string     `gorm:"column:channel"`       // email | sms | whatsapp
	Provider     string     `gorm:"column:provider"`      // mailtrap | sendgrid | etc.
	Status       string     `gorm:"column:status"`        // pending | sent | failed
	SentAt       *time.Time `gorm:"column:sent_at"`
	DeliveredAt  *time.Time `gorm:"column:delivered_at"`
	ErrorMessage string     `gorm:"column:error_message"`
	Otp          Otp        `gorm:"foreignKey:OtpID;references:ID"`
}

func (o *OtpDelivery) TableName() string {
	return "otp_deliveries"
}
