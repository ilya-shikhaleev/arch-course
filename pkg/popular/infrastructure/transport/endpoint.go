package transport

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/ilya-shikhaleev/arch-course/pkg/popular/app/popular"
)

type readProductsRequest struct {
	Count int
}

type readProductsResponse struct {
	Products []product `json:"products,omitempty"`
}

type product struct {
	ProductID   string  `json:"productID,omitempty"`
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Material    string  `json:"material,omitempty"`
	Height      *int    `json:"height,omitempty"`
	Color       *string `json:"color,omitempty"`
	Price       float32 `json:"price,omitempty"`
}

func makeReadProductsEndpoint(repo popular.Repository) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(readProductsRequest)
		if products, err := repo.FindPopular(req.Count); err != nil {
			return readProductsResponse{}, err
		} else {
			var responseProducts []product
			for _, p := range products {
				responseProduct := product{
					ProductID:   p.ID,
					Title:       p.Title,
					Description: p.Description,
					Material:    p.Material,
					Height:      p.Height,
					Color:       p.Color,
					Price:       p.Price,
				}
				responseProducts = append(responseProducts, responseProduct)
			}

			return readProductsResponse{Products: responseProducts}, nil
		}
	}
}
