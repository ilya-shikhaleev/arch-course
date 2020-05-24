package user

type ID string
type Email string
type Phone string

type Repository interface {
	Store(*User) error
	Find(ID) (*User, error)
	FindByUsername(string) (*User, error)
	Remove(ID) error
	NextID() (ID, error)
}

type User struct {
	ID          ID
	Username    string
	FirstName   string
	LastName    string
	Email       Email
	Phone       Phone
	EncodedPass string
}
