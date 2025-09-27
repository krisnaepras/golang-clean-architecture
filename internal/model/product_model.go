package model

type ProductResponse struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Price       int64  `json:"price,omitempty"`
	Stock       int    `json:"stock,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	CreatedAt   int64  `json:"created_at,omitempty"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

type CreateProductRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=500"`
	Price       int64  `json:"price" validate:"required,min=0"`
	Stock       int    `json:"stock" validate:"required,min=0"`
}

type UpdateProductRequest struct {
	ID          string `json:"-" validate:"required"`
	Name        string `json:"name,omitempty" validate:"max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	Price       int64  `json:"price,omitempty" validate:"min=0"`
	Stock       int    `json:"stock,omitempty" validate:"min=0"`
}

type GetProductRequest struct {
	ID string `validate:"required"`
}

type SearchProductRequest struct {
	Name        string `json:"name"`
	PriceMin    int64  `json:"price_min"`
	PriceMax    int64  `json:"price_max"`
	Page        int    `json:"page" validate:"min=1"`
	Size        int    `json:"size" validate:"min=1,max=100"`
}