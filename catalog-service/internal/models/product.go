package models

// ProductRequest is used for creating a product
type ProductRequest struct {
	Name       string  `json:"name" validate:"required"`
	Price      float64 `json:"price" validate:"required"`
	CategoryID int64   `json:"category_id" validate:"required"`
}

// ProductUpdateRequest allows partial updates
type ProductUpdateRequest struct {
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	CategoryID int64   `json:"category_id"`
}

// ProductResponse is returned to the client
type ProductResponse struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	CategoryID int64   `json:"category_id"`
}
