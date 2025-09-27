package converter

import (
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"
)

func ProductToResponse(product *entity.Product) *model.ProductResponse {
	return &model.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		UserID:      product.UserID,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func ProductsToResponses(products []entity.Product) []model.ProductResponse {
	var responses []model.ProductResponse
	for _, product := range products {
		responses = append(responses, *ProductToResponse(&product))
	}
	return responses
}