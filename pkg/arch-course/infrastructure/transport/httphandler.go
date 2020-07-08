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

	"github.com/ilya-shikhaleev/arch-course/pkg/arch-course/app/user"
)

type userInfo struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
}

func MakeHandler(s *user.Service, logger httplog.Logger) http.Handler {
	r := mux.NewRouter()
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	createUserHandler := httptransport.NewServer(
		makeCreateUserEndpoint(s),
		decodeCreateUserRequest,
		encodeResponse,
		opts...,
	)

	readUserHandler := httptransport.NewServer(
		makeReadUserEndpoint(s),
		decodeReadUserRequest,
		encodeResponse,
		opts...,
	)

	updateUserHandler := httptransport.NewServer(
		makeUpdateUserEndpoint(s),
		decodeUpdateUserRequest,
		encodeResponse,
		opts...,
	)

	deleteUserHandler := httptransport.NewServer(
		makeDeleteUserEndpoint(s),
		decodeRemoveUserRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/users", createUserHandler).Methods(http.MethodPost)
	r.Handle("/api/v1/users/{id}", readUserHandler).Methods(http.MethodGet)
	r.Handle("/api/v1/users/{id}", updateUserHandler).Methods(http.MethodPut)
	r.Handle("/api/v1/user", updateUserHandler).Methods(http.MethodPut)
	r.Handle("/api/v1/users/{id}", deleteUserHandler).Methods(http.MethodDelete)

	return r
}

func decodeCreateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var info userInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		return nil, newErrInvalidRequest(err, "invalid create user request")
	}

	req := createUserRequest{
		Username:  info.Username,
		FirstName: info.FirstName,
		LastName:  info.LastName,
		Email:     info.Email,
		Phone:     info.Phone,
		Password:  info.Password,
	}

	return req, nil
}

func decodeReadUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, newErrInvalidRequest(nil, "id required for read user request")
	}

	if r.Header.Get("X-User-Id") != id {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user data (%s != %s)", id, r.Header.Get("X-User-Id")))
	}

	req := readUserRequest{UserID: id}
	return req, nil
}

func decodeUpdateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		if r.Header.Get("X-User-Id") == "" {
			return nil, newErrInvalidRequest(nil, "id required for update user request")
		} else {
			id = r.Header.Get("X-User-Id")
		}
	}
	if r.Header.Get("X-User-Id") != id {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user data (%s != %s)", id, r.Header.Get("X-User-Id")))
	}

	var info userInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		return nil, newErrInvalidRequest(err, "invalid update user request")
	}

	req := updateUserRequest{
		UserID:    id,
		Username:  info.Username,
		FirstName: info.FirstName,
		LastName:  info.LastName,
		Email:     info.Email,
		Phone:     info.Phone,
	}
	return req, nil
}

func decodeRemoveUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, newErrInvalidRequest(nil, "id required for delete user request")
	}

	if r.Header.Get("X-User-Id") != id {
		return nil, newErrUnauthorized(fmt.Sprintf("can read only self user data (%s != %s)", id, r.Header.Get("X-User-Id")))
	}

	req := deleteUserRequest{UserID: id}
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
		case user.ErrUserNotFound:
			w.WriteHeader(http.StatusNotFound)
		case user.ErrDuplicateUsername:
			w.WriteHeader(http.StatusBadRequest)
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
