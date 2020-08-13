package transport

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/ilya-shikhaleev/arch-course/pkg/cart/app/cart"
)

type readCartRequest struct {
	UserID string
}

type readCartResponse struct {
	CartID     string   `json:"cartID,omitempty"`
	ProductIDs []string `json:"productIDs,omitempty"`
}

func makeReadCartEndpoint(repo cart.Repository) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(readCartRequest)
		if c, err := repo.FindByUserID(req.UserID); err != nil {
			return readCartResponse{}, err
		} else {
			return readCartResponse{
				CartID:     string(c.ID),
				ProductIDs: c.ProductIDs,
			}, nil
		}
	}
}

type addProductToCartRequest struct {
	UserID    string
	ProductID string
}

type addProductToCartResponse struct {
}

func makeAddProductToCartEndpoint(service *cart.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addProductToCartRequest)
		if err := service.AddProductToCart(req.UserID, req.ProductID); err != nil {
			return addProductToCartResponse{}, err
		} else {
			return addProductToCartResponse{}, nil
		}
	}
}
