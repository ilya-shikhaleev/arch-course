package transport

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/ilya-shikhaleev/arch-course/pkg/product/app/product"
)

type readProductRequest struct {
	ID string
}

type readProductResponse struct {
	MetaProductID string  `json:"metaProductID,omitempty"`
	Title         string  `json:"title,omitempty"`
	Description   string  `json:"description,omitempty"`
	Material      int     `json:"material,omitempty"`
	Height        *int    `json:"height,omitempty"`
	Color         *string `json:"color,omitempty"`
	Price         float32 `json:"price,omitempty"`
}

func makeReadProductEndpoint(repo product.Repository) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(readProductRequest)
		if p, err := repo.FindByID(product.ID(req.ID)); err != nil {
			return readProductResponse{}, err
		} else {
			return readProductResponse{
				MetaProductID: string(p.MetaProductID),
				Title:         p.Title,
				Description:   p.Description,
				Material:      int(p.Material),
				Height:        p.Height,
				Color:         p.Color,
				Price:         p.Price,
			}, nil
		}
	}
}
