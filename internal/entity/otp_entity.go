package entity

import "time"

type Otp struct {
	ID           string     `gorm:"column:id;primaryKey"`
	UserID       string     `gorm:"column:user_id"`
	Destination  string     `gorm:"column:destination"`
	OtpCode      string     `gorm:"column:otp_code"`
	Purpose      string     `gorm:"column:purpose"`
	ExpiredAt    *time.Time `gorm:"column:expired_at"`
	VerifiedAt   *time.Time `gorm:"column:verified_at"`
	AttemptCount int        `gorm:"column:attempt_count;default:0"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	User         User       `gorm:"foreignKey:UserID;references:ID"`
}

func (o *Otp) TableName() string {
	return "otps"
}
