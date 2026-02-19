package entity

import "time"

type OauthState struct {
	ID          string    `gorm:"column:id;primaryKey"`
	State       string    `gorm:"column:state"`
	Provider    string    `gorm:"column:provider"`
	RedirectURI string    `gorm:"column:redirect_uri"`
	ExpiredAt   time.Time `gorm:"column:expired_at"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (o *OauthState) TableName() string {
	return "oauth_states"
}
