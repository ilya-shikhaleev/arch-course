package user

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
	repo := &mockRepo{}
	service := NewService(repo, mockEncoder())

	userID, err := service.CreateUser("username", "first name", "last name", "some@email.ru", "900", "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(repo.users))
	user, err := repo.Find(userID)
	assert.Nil(t, err)
	assert.NotNil(t, 1, user)

	_, err = service.CreateUser("username", "first name", "last name", "some@email.ru", "900", "")
	assert.Equal(t, ErrDuplicateUsername, err)
}

func TestUserService_UpdateUser(t *testing.T) {
	repo := &mockRepo{}
	service := NewService(repo, mockEncoder())

	userID, err := service.CreateUser("username", "first name", "last name", "some@email.ru", "900", "")
	assert.Nil(t, err)
	_, err = service.CreateUser("username2", "first name2", "last name2", "some@email.ru", "900", "")
	assert.Nil(t, err)

	err = service.UpdateUser(userID, "username2", "first name", "last name", "some@email.ru", "900")
	assert.Equal(t, ErrDuplicateUsername, err)

	updatedUserName := "username3"
	err = service.UpdateUser(userID, updatedUserName, "first name3", "last name3", "some@email.ru", "900")
	assert.Nil(t, err)
	user, err := repo.Find(userID)
	assert.Nil(t, err)
	assert.Equal(t, updatedUserName, user.Username)
}

type mockRepo struct {
	users []*User
}

func (repo *mockRepo) Store(user *User) error {
	for _, u := range repo.users {
		if u.ID == user.ID {
			return nil
		}
	}

	repo.users = append(repo.users, user)
	return nil
}

func (repo *mockRepo) Find(id ID) (*User, error) {
	for _, u := range repo.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (repo *mockRepo) FindByUsername(username string) (*User, error) {
	for _, u := range repo.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (repo *mockRepo) Remove(id ID) error {
	for i, u := range repo.users {
		if u.ID == id {
			repo.users = append(repo.users[:i], repo.users[i+1:]...)
			return nil
		}
	}
	return ErrUserNotFound
}

func (repo *mockRepo) NextID() (ID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return ID(id.String()), nil
}

func mockEncoder() EncoderFunc {
	return func(s string) string {
		return s
	}
}
