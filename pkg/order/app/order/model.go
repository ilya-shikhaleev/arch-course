package order

import (
	"errors"
)

type ID string

type Status int

const (
	PendingPayment Status = iota
	Completed
)

type Order struct {
	ID       ID
	UserID   string
	Status   Status
	Products []Product
}

type Product struct {
	ProductID string
	Price     float32
}

func (o Order) Price() float32 {
	result := float32(0.0)
	for _, p := range o.Products {
		result += p.Price
	}
	return result
}

type Repository interface {
	FindByID(ID) (*Order, error)
	FindByUserID(userID string) ([]Order, error)
	Store(*Order) error
	NextID() (ID, error)
}

var ErrOrderNotFound = errors.New("order not found")
var ErrEmptyCart = errors.New("cart is empty")
