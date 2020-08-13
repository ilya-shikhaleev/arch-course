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

	"github.com/ilya-shikhaleev/arch-course/pkg/cart/app/cart"
)

func MakeHandler(service *cart.Service, repo cart.Repository, logger httplog.Logger) http.Handler {
	r := mux.NewRouter()
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	readCartHandler := httptransport.NewServer(
		makeReadCartEndpoint(repo),
		decodeReadCartRequest,
		encodeResponse,
		opts...,
	)

	addProductToCartHandler := httptransport.NewServer(
		makeAddProductToCartEndpoint(service),
		decodeAddProductToCartRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/cart", readCartHandler).Methods(http.MethodGet)
	r.Handle("/api/v1/cart/product", addProductToCartHandler).Methods(http.MethodPut)

	return r
}

func decodeReadCartRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user cart (%s)", r.Header.Get("X-User-Id")))
	}

	req := readCartRequest{UserID: userID}
	return req, nil
}

type addProductToCartRequestBody struct {
	ProductID string `json:"productID"`
}

func decodeAddProductToCartRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user cart (%s)", r.Header.Get("X-User-Id")))
	}

	var body addProductToCartRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, newErrInvalidRequest(err, "invalid add product request")
	}
	req := addProductToCartRequest{UserID: userID, ProductID: body.ProductID}
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
		case cart.ErrCartNotFound:
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
