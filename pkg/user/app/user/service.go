package user

import (
	"errors"
)

func NewService(repo Repository, passEncoder PassEncoder) *Service {
	return &Service{repo, passEncoder}
}

type Service struct {
	repo        Repository
	passEncoder PassEncoder
}

func (s *Service) CreateUser(username, firstName, lastName string, email Email, phone Phone, password string) (userID ID, err error) {
	if u, err := s.repo.FindByUsername(username); u != nil {
		return userID, ErrDuplicateUsername
	} else if err != ErrUserNotFound {
		return userID, err
	}

	userID, err = s.repo.NextID()
	if err != nil {
		return "", err
	}
	err = s.repo.Store(&User{
		ID:          userID,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		Phone:       phone,
		EncodedPass: s.passEncoder.Encode(password),
	})

	return userID, err
}

func (s *Service) UpdateUser(id ID, username, firstName, lastName string, email Email, phone Phone) error {
	user, err := s.repo.Find(id)
	if err != nil {
		return err
	}
	if user2, err := s.repo.FindByUsername(username); user2 != nil && user2.ID != id {
		return ErrDuplicateUsername
	} else if err != nil && err != ErrUserNotFound {
		return err
	}

	user.Username = username
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Phone = phone

	return s.repo.Store(user)
}

func (s *Service) DeleteUser(id ID) error {
	return s.repo.Remove(id)
}

func (s *Service) ReadUser(id ID) (*User, error) {
	return s.repo.Find(id)
}

var ErrUserNotFound = errors.New("user not found")
var ErrDuplicateUsername = errors.New("duplicate username")
