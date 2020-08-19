package transport

import (
	"context"
	"encoding/json"
	"net/http"

	httplog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/ilya-shikhaleev/arch-course/pkg/popular/app/popular"
)

func MakeHandler(repo popular.Repository, logger httplog.Logger) http.Handler {
	r := mux.NewRouter()
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	readProductsHandler := httptransport.NewServer(
		makeReadProductsEndpoint(repo),
		decodeReadProductsRequest,
		encodeResponse,
		opts...,
	)

	onBuyProductsHandler := httptransport.NewServer(
		makeOnBuyProductsEndpoint(repo),
		decodeOnBuyProductsRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/popular", readProductsHandler).Methods(http.MethodGet)
	r.Handle("/api/v1/internal/popular/buy", onBuyProductsHandler).Methods(http.MethodPost)

	return r
}

func decodeReadProductsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Count int `json:"count,omitempty"`
	}
	count := 6 // default value
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
		if body.Count > 0 {
			count = body.Count
		}
	}

	req := readProductsRequest{Count: count}
	return req, nil
}

func decodeOnBuyProductsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		ProductIDs []string `json:"productIDs,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, newErrInvalidRequest(err, "invalid on buy product request")
	}
	req := onBuyProductsRequest{ProductIDs: body.ProductIDs}
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
		case popular.ErrProductNotFound:
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
