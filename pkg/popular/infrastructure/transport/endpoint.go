package transport

import (
	"context"
	"encoding/json"
	"net/http"

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

type onBuyProductsRequest struct {
	ProductIDs []string
}

type onBuyProductsResponse struct {
}

func makeOnBuyProductsEndpoint(repo popular.Repository) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(onBuyProductsRequest)

		var products []product
		for _, productID := range req.ProductIDs {
			p, err := getProduct(productID)
			if err != nil {
				return onBuyProductsResponse{}, err
			}
			products = append(products, p)
		}

		for _, p := range products {
			product, err := repo.FindByID(p.ProductID)
			if err == popular.ErrProductNotFound {
				product = &popular.Product{
					ID:          p.ProductID,
					Title:       p.Title,
					Description: p.Description,
					Material:    p.Material,
					Height:      p.Height,
					Color:       p.Color,
					Price:       p.Price,
					BuyCount:    0,
				}
			} else if err != nil {
				return onBuyProductsResponse{}, err
			}

			product.BuyCount = product.BuyCount + 1
			err = repo.Store(product)
			if err != nil {
				return onBuyProductsResponse{}, err
			}
		}

		return onBuyProductsResponse{}, nil
	}
}

func getProduct(productID string) (product, error) {
	const cartHost = "http://product-product-chart.arch-course.svc.cluster.local:9000" // TODO: use env variable here
	req, err := http.NewRequest(http.MethodGet, cartHost+"/api/v1/internal/products/"+productID, nil)
	if err != nil {
		return product{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return product{}, err
	}

	var p struct {
		Title       string  `json:"title,omitempty"`
		Description string  `json:"description,omitempty"`
		Material    string  `json:"material,omitempty"`
		Height      *int    `json:"height,omitempty"`
		Color       *string `json:"color,omitempty"`
		Price       float32 `json:"price,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&p)
	if err != nil {
		return product{}, err
	}
	return product{
		ProductID:   productID,
		Title:       p.Title,
		Description: p.Description,
		Material:    p.Material,
		Height:      p.Height,
		Color:       p.Color,
		Price:       p.Price,
	}, nil
}
