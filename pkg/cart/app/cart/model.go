package cart

import (
	"errors"
)

type ID string

type Cart struct {
	ID         ID
	UserID     string
	ProductIDs []string
}

type Repository interface {
	FindByID(id ID) (*Cart, error)
	FindByUserID(userID string) (*Cart, error)
	Store(cart *Cart) error
	NextID() (ID, error)
}

var ErrCartNotFound = errors.New("cart not found")
