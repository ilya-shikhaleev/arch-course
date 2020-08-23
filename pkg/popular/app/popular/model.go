package popular

import "errors"

type Product struct {
	ID          string  `json:"id,omitempty"`
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Material    string  `json:"material,omitempty"`
	Height      *int    `json:"height,omitempty"`
	Color       *string `json:"color,omitempty"`
	Price       float32 `json:"price,omitempty"`
	BuyCount    int     `json:"buyCount,omitempty"`
}

type Repository interface {
	FindByID(id string) (*Product, error)
	FindPopular(count int) ([]Product, error)
	Store(product *Product) error
}

var ErrProductNotFound = errors.New("product not found")
