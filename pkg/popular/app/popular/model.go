package popular

import "errors"

type Product struct {
	ID          string
	Title       string
	Description string
	Material    string
	Height      *int
	Color       *string
	Price       float32
	BuyCount    int
}

type Repository interface {
	FindByID(id string) (*Product, error)
	FindPopular(count int) ([]Product, error)
	Store(product *Product) error
}

var ErrProductNotFound = errors.New("product not found")
