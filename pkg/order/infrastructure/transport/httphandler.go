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

	"github.com/ilya-shikhaleev/arch-course/pkg/order/app/order"
)

func MakeHandler(service *order.Service, repo order.Repository, logger httplog.Logger) http.Handler {
	r := mux.NewRouter()
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	readOrderHandler := httptransport.NewServer(
		makeReadOrdersEndpoint(repo),
		decodeReadOrdersRequest,
		encodeResponse,
		opts...,
	)

	createOrderHandler := httptransport.NewServer(
		makeCreateOrderEndpoint(service),
		decodeCreateOrderRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/orders", readOrderHandler).Methods(http.MethodGet)
	r.Handle("/api/v1/orders", createOrderHandler).Methods(http.MethodPost)

	return r
}

func decodeReadOrdersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user order (%s)", r.Header.Get("X-User-Id")))
	}

	req := readOrdersRequest{UserID: userID}
	return req, nil
}

func decodeCreateOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user order (%s)", r.Header.Get("X-User-Id")))
	}
	req := createOrderRequest{UserID: userID}
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
		switch err {
		case order.ErrOrderNotFound:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
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
