package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilya-shikhaleev/arch-course/pkg/popular/app/popular"
)

type OnBuyProductsRequest struct {
	ProductIDs []string `json:"productIDs,omitempty"`
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

func OnBuyProducts(req OnBuyProductsRequest, repo popular.Repository) error {
	var products []product
	for _, productID := range req.ProductIDs {
		p, err := getProduct(productID)
		if err != nil {
			return err
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
			return err
		}

		product.BuyCount = product.BuyCount + 1
		err = repo.Store(product)
		if err != nil {
			return err
		}
	}

	return nil
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
