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
	Material      string  `json:"material,omitempty"`
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
				Material:      materialString(p.Material),
				Height:        p.Height,
				Color:         p.Color,
				Price:         p.Price,
			}, nil
		}
	}
}

type readProductsRequest struct {
	Search string
}

type readProductsResponse struct {
	Products []responseProduct `json:"products,omitempty"`
}

type responseProduct struct {
	ProductID     string  `json:"productID,omitempty"`
	MetaProductID string  `json:"metaProductID,omitempty"`
	Title         string  `json:"title,omitempty"`
	Description   string  `json:"description,omitempty"`
	Material      string  `json:"material,omitempty"`
	Height        *int    `json:"height,omitempty"`
	Color         *string `json:"color,omitempty"`
	Price         float32 `json:"price,omitempty"`
}

func makeReadProductsEndpoint(repo product.Repository) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(readProductsRequest)
		if products, err := repo.FindBySpecification(product.Specification{SearchString: req.Search}); err != nil {
			return readProductsResponse{}, err
		} else {
			var responseProducts []responseProduct
			for _, p := range products {
				responseProduct := responseProduct{
					ProductID:     string(p.ID),
					MetaProductID: string(p.MetaProductID),
					Title:         p.Title,
					Description:   p.Description,
					Material:      materialString(p.Material),
					Height:        p.Height,
					Color:         p.Color,
					Price:         p.Price,
				}
				responseProducts = append(responseProducts, responseProduct)
			}

			return readProductsResponse{Products: responseProducts}, nil
		}
	}
}

func materialString(materialType product.Material) string {
	var material string
	switch materialType {
	case product.Paper:
		material = "paper"
	case product.FullMetal:
		material = "fullmetal"
	case product.Template:
		material = "template"
	}
	return material
}
