package entity

import "time"

type RefreshToken struct {
	ID         string    `gorm:"column:id;primaryKey"`
	UserID     string    `gorm:"column:user_id"`
	TokenHash  string    `gorm:"column:token_hash"`
	DeviceInfo string    `gorm:"column:device_info"`
	IPAddress  string    `gorm:"column:ip_address"`
	IsRevoked  bool      `gorm:"column:is_revoked;default:false"`
	ExpiredAt  time.Time `gorm:"column:expired_at"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
	User       User      `gorm:"foreignKey:UserID;references:ID"`
}

func (r *RefreshToken) TableName() string {
	return "refresh_tokens"
}
