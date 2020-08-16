package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httplog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func MakeHandler(logger httplog.Logger) http.Handler {
	r := mux.NewRouter()
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	payOrderHandler := httptransport.NewServer(
		makePayOrderEndpoint(),
		decodePayOrderRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/payment", payOrderHandler).Methods(http.MethodPost)

	return r
}

func decodePayOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user order (%s)", r.Header.Get("X-User-Id")))
	}

	var body struct {
		OrderID string `json:"orderID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, newErrInvalidRequest(err, "invalid pay order request")
	}

	req := payOrderRequest{UserID: userID, OrderID: body.OrderID}
	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if invalidRequestErr, ok := err.(*errInvalidRequest); ok {
		w.WriteHeader(http.StatusBadRequest)
		err = errors.New(invalidRequestErr.message)
	} else if unauthorizedErr, ok := err.(*errUnauthorized); ok {
		w.WriteHeader(http.StatusUnauthorized)
		err = errors.New(unauthorizedErr.message)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

type errorer interface {
	error() error
}

type errInvalidRequest struct {
	message string
	orig    error
}

func newErrInvalidRequest(err error, message string) *errInvalidRequest {
	return &errInvalidRequest{
		message: message,
		orig:    err,
	}
}

func (e *errInvalidRequest) Error() string {
	if e.orig == nil {
		return e.message
	}
	return errors.Wrap(e.orig, e.message).Error()
}

type errUnauthorized struct {
	message string
}

func newErrUnauthorized(message string) *errUnauthorized {
	return &errUnauthorized{
		message: message,
	}
}

func (e *errUnauthorized) Error() string {
	return e.message
}
