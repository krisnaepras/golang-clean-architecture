package entity

import "time"

type User struct {
	ID            string     `gorm:"column:id;primaryKey"`
	FullName      string     `gorm:"column:fullname"`
	Email         string     `gorm:"column:email"`
	Phone         string     `gorm:"column:phone"`
	BirthDate     *time.Time `gorm:"column:birth_date"`
	Address       string     `gorm:"column:address"`
	SubdistrictID *int       `gorm:"column:subdistrict_id"`
	DistrictID    *int       `gorm:"column:district_id"`
	CityID        *int       `gorm:"column:city_id"`
	ProfileImage  string     `gorm:"column:profile_image"`
	IsActive      bool       `gorm:"column:is_active;default:true"`
	EmailVerified bool       `gorm:"column:email_verified;default:false"`
	CreatedBy     string     `gorm:"column:created_by"`
	UpdatedBy     string     `gorm:"column:updated_by"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

func (u *User) TableName() string {
	return "users"
}
