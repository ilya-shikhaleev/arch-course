package transport

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/ilya-shikhaleev/arch-course/pkg/order/app/order"
)

func NewProductsRetriever() *Retriever {
	return &Retriever{}
}

type Retriever struct {
}

func (r *Retriever) OrderProducts(userID string) ([]order.Product, error) {
	const cartHost = "http://cart-cart-chart.arch-course.svc.cluster.local:9000" // TODO: use env variable here
	req, err := http.NewRequest(http.MethodGet, cartHost+"/api/v1/cart", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-User-Id", userID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var readCartResponse struct {
		CartID     string   `json:"cartID,omitempty"`
		ProductIDs []string `json:"productIDs,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&readCartResponse)
	if err != nil {
		return nil, err
	}

	var products []order.Product
	for _, productID := range readCartResponse.ProductIDs {
		price, err := retrieveProductPrice(userID, productID)
		if err != nil {
			return nil, err
		}
		products = append(products, order.Product{
			ProductID: productID,
			Price:     price,
		})
	}

	return products, clearCart(userID)
}

func (r *Retriever) RestoreProducts(userID string, products []order.Product) error {
	const cartHost = "http://cart-cart-chart.arch-course.svc.cluster.local:9000" // TODO: use env variable here

	for _, product := range products {
		var jsonStr = []byte(`{"productID":"` + product.ProductID + `"}`)
		req, err := http.NewRequest(http.MethodPut, cartHost+"/api/v1/cart/product", bytes.NewBuffer(jsonStr))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("X-User-Id", userID)
		client := &http.Client{}
		_, err = client.Do(req)
		if err != nil {
			return err
		}
	}

	return nil
}

func retrieveProductPrice(userID, productID string) (float32, error) {
	const host = "http://product-product-chart.arch-course.svc.cluster.local:9000" // TODO: use env variable here
	req, err := http.NewRequest(http.MethodGet, host+"/api/v1/products/"+productID, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("X-User-Id", userID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	var readProductResponse struct {
		Price float32 `json:"price,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&readProductResponse)
	if err != nil {
		return 0, err
	}
	return readProductResponse.Price, nil
}

func clearCart(userID string) error {
	const cartHost = "http://cart-cart-chart.arch-course.svc.cluster.local:9000" // TODO: use env variable here
	req, err := http.NewRequest(http.MethodDelete, cartHost+"/api/v1/cart", nil)
	if err != nil {
		return err
	}

	req.Header.Add("X-User-Id", userID)
	client := &http.Client{}
	_, err = client.Do(req)
	return err
}
