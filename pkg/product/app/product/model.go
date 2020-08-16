package product

import "errors"

type ID string
type MetaProductID string
type Material int

const (
	Paper     Material = 1
	FullMetal Material = 2
	Template  Material = 3
)

type Product struct {
	ID            ID
	MetaProductID MetaProductID
	Title         string
	Description   string
	Material      Material
	Height        *int
	Color         *string
	Price         float32
}

type Repository interface {
	FindByID(id ID) (*Product, error)
	FindBySpecification(specification Specification) ([]Product, error)
}

type Specification struct {
	SearchString string
}

var ErrProductNotFound = errors.New("product not found")
