package transport

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/ilya-shikhaleev/arch-course/pkg/arch-course/app/user"
)

type createUserRequest struct {
	Username  string
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

type createUserResponse struct {
	UserID string `json:"userId,omitempty"`
}

func makeCreateUserEndpoint(s user.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createUserRequest)
		userID, err := s.CreateUser(
			req.Username,
			req.FirstName,
			req.LastName,
			user.Email(req.Email),
			user.Phone(req.Phone),
		)
		return createUserResponse{UserID: string(userID)}, err
	}
}

type readUserRequest struct {
	UserID string
}

type readUserResponse struct {
	Username  string `json:"username,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

func makeReadUserEndpoint(s user.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(readUserRequest)
		if u, err := s.ReadUser(user.ID(req.UserID)); err != nil {
			return readUserResponse{}, err
		} else {
			return readUserResponse{
				Username:  u.Username,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Email:     string(u.Email),
				Phone:     string(u.Phone),
			}, nil
		}
	}
}

type updateUserRequest struct {
	UserID    string
	Username  string
	FirstName string
	LastName  string
	Email     string
	Phone     string
}
type updateUserResponse struct {
	Err error `json:"error,omitempty"`
}

func makeUpdateUserEndpoint(s user.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateUserRequest)
		err := s.UpdateUser(
			user.ID(req.UserID),
			req.Username,
			req.FirstName,
			req.LastName,
			user.Email(req.Email),
			user.Phone(req.Phone),
		)
		return updateUserResponse{Err: err}, nil
	}
}

type deleteUserRequest struct {
	UserID string
}

type deleteUserResponse struct {
	Err error `json:"error,omitempty"`
}

func makeDeleteUserEndpoint(s user.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteUserRequest)
		err := s.DeleteUser(user.ID(req.UserID))
		return deleteUserResponse{Err: err}, nil
	}
}
