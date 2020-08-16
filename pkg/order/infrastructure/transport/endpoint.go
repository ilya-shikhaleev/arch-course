package transport

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/ilya-shikhaleev/arch-course/pkg/order/app/order"
)

type readOrdersRequest struct {
	UserID string
}

type readOrdersResponse struct {
	Orders []Order `json:"orders,omitempty"`
}

type Order struct {
	OrderID  string    `json:"orderID,omitempty"`
	Status   string    `json:"status,omitempty"`
	Products []Product `json:"products,omitempty"`
}

type Product struct {
	Price     float32 `json:"price,omitempty"`
	ProductID string  `json:"productID,omitempty"`
}

func makeReadOrdersEndpoint(repo order.Repository) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(readOrdersRequest)
		if orders, err := repo.FindByUserID(req.UserID); err != nil {
			return readOrdersResponse{}, err
		} else {
			var responseOrders []Order
			for _, o := range orders {
				var status string
				switch o.Status {
				case order.PendingPayment:
					status = "pending"
				case order.Completed:
					status = "completed"
				}

				var products []Product
				for _, p := range o.Products {
					products = append(products, Product{
						Price:     p.Price,
						ProductID: p.ProductID,
					})
				}
				responseOrders = append(responseOrders, Order{
					OrderID:  string(o.ID),
					Status:   status,
					Products: products,
				})
			}

			return readOrdersResponse{Orders: responseOrders}, nil
		}
	}
}

type createOrderRequest struct {
	UserID string
}

type createOrderResponse struct {
	OrderID string `json:"orderID,omitempty"`
}

func makeCreateOrderEndpoint(service *order.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createOrderRequest)
		if id, err := service.CreateOrder(req.UserID); err != nil {
			return createOrderResponse{}, err
		} else {
			return createOrderResponse{OrderID: string(id)}, nil
		}
	}
}
