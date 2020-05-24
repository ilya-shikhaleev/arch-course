package postgres

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/ilya-shikhaleev/arch-course/pkg/arch-course/app/user"
)

func NewUserRepository(db *sql.DB) user.Repository {
	return &userRepository{db: db}
}

type userRepository struct {
	db *sql.DB
}

func (repo *userRepository) Store(user *user.User) error {
	sqlStatement := `
		INSERT INTO users (id, username, firstname, lastname, email, phone, password)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET username  = EXCLUDED.username,
									   firstname = EXCLUDED.firstname,
									   lastname  = EXCLUDED.lastname,
									   email     = EXCLUDED.email,
									   phone     = EXCLUDED.phone;
`
	_, err := repo.db.Exec(sqlStatement, string(user.ID), user.Username, user.FirstName, user.LastName, string(user.Email), string(user.Phone), user.EncodedPass)
	return err
}

func (repo *userRepository) Find(id user.ID) (*user.User, error) {
	sqlStatement := `SELECT id, username, firstname, lastname, email, phone, password FROM users WHERE id=$1;`
	var u user.User
	row := repo.db.QueryRow(sqlStatement, string(id))
	switch err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Phone, &u.EncodedPass); err {
	case sql.ErrNoRows:
		return nil, user.ErrUserNotFound
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func (repo *userRepository) FindByUsername(username string) (*user.User, error) {
	sqlStatement := `SELECT id, username, firstname, lastname, email, phone, password FROM users WHERE username=$1;`
	var u user.User
	row := repo.db.QueryRow(sqlStatement, string(username))
	switch err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Phone, &u.EncodedPass); err {
	case sql.ErrNoRows:
		return nil, user.ErrUserNotFound
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func (repo *userRepository) Remove(id user.ID) error {
	sqlStatement := `
		DELETE FROM users
		WHERE id = $1;`
	_, err := repo.db.Exec(sqlStatement, string(id))
	return err
}

func (repo *userRepository) NextID() (user.ID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return user.ID(id.String()), nil
}
