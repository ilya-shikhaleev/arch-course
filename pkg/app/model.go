package app

type UserID string
type Email string
type Phone string

type UserRepository interface {
	Store(*User) error
	Find(UserID) (*User, error)
	FindByUsername(string) (*User, error)
	Remove(UserID) error
	NextID() (UserID, error)
}

type User struct {
	ID        UserID
	Username  string
	FirstName string
	LastName  string
	Email     Email
	Phone     Phone
}
