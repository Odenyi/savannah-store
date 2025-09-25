package models

type CartItem struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id" validate:"required"`
	ProductID int64   `json:"product_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required"`
	Price     float64 `json:"price"`
}

type Order struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Total     float64    `json:"total"`
	Status    string     `json:"status"`
	Items     []CartItem `json:"items"`
	CreatedAt string     `json:"created_at"`
}
type AddToCartRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type UpdateCartRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type PlaceOrderRequest struct {
	Address       string `json:"address"`
	PaymentMethod string `json:"payment_method"`
}
