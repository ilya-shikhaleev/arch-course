package transport

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

type payOrderRequest struct {
	UserID  string
	OrderID string
}

type payOrderResponse struct {
}

func makePayOrderEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(payOrderRequest)

		const host = "http://order-order-chart.arch-course.svc.cluster.local:9000" // TODO: use env variable here
		httpReq, err := http.NewRequest(http.MethodPatch, host+"/api/v1/internal/orders/"+req.OrderID, nil)
		if err != nil {
			return 0, err
		}

		httpReq.Header.Add("X-User-Id", req.UserID)
		client := &http.Client{}
		_, err = client.Do(httpReq)
		return payOrderResponse{}, err
	}
}
