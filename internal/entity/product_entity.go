package entity

// Product adalah struct yang merepresentasikan entitas produk
type Product struct {
	ID          string `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Price       int64  `gorm:"column:price"`
	Stock       int    `gorm:"column:stock"`
	UserID      string `gorm:"column:user_id"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	User        User   `gorm:"foreignKey:user_id;references:id"`
}

func (p *Product) TableName() string {
	return "products"
}