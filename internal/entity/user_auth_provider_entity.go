package entity

import "time"

type UserAuthProvider struct {
	ID             string    `gorm:"column:id;primaryKey"`
	UserID         string    `gorm:"column:user_id"`
	Provider       string    `gorm:"column:provider"`
	ProviderUserID string    `gorm:"column:provider_user_id"`
	Email          string    `gorm:"column:email"`
	PasswordHash   string    `gorm:"column:password_hash"`
	CreatedBy      string    `gorm:"column:created_by"`
	UpdatedBy      string    `gorm:"column:updated_by"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
	User           User      `gorm:"foreignKey:UserID;references:ID"`
}

func (u *UserAuthProvider) TableName() string {
	return "user_auth_providers"
}
