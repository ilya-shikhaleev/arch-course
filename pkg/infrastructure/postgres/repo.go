package postgres

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/ilya-shikhaleev/arch-course/pkg/app"
)

func NewUserRepository(db *sql.DB) app.UserRepository {
	return &userRepository{db: db}
}

type userRepository struct {
	db *sql.DB
}

func (repo *userRepository) Store(user *app.User) error {
	sqlStatement := `
		INSERT INTO users (id, username, firstname, lastname, email, phone)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET username  = EXCLUDED.username,
									   firstname = EXCLUDED.firstname,
									   lastname  = EXCLUDED.lastname,
									   email     = EXCLUDED.email,
									   phone     = EXCLUDED.phone;
`
	_, err := repo.db.Exec(sqlStatement, string(user.ID), user.Username, user.FirstName, user.LastName, string(user.Email), string(user.Phone))
	return err
}

func (repo *userRepository) Find(id app.UserID) (*app.User, error) {
	sqlStatement := `SELECT id, username, firstname, lastname, email, phone FROM users WHERE id=$1;`
	var user app.User
	row := repo.db.QueryRow(sqlStatement, string(id))
	switch err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Phone); err {
	case sql.ErrNoRows:
		return nil, app.ErrUserNotFound
	case nil:
		return &user, nil
	default:
		return nil, err
	}
}

func (repo *userRepository) FindByUsername(username string) (*app.User, error) {
	sqlStatement := `SELECT id, username, firstname, lastname, email, phone FROM users WHERE username=$1;`
	var user app.User
	row := repo.db.QueryRow(sqlStatement, string(username))
	switch err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Phone); err {
	case sql.ErrNoRows:
		return nil, app.ErrUserNotFound
	case nil:
		return &user, nil
	default:
		return nil, err
	}
}

func (repo *userRepository) Remove(id app.UserID) error {
	sqlStatement := `
		DELETE FROM users
		WHERE id = $1;`
	_, err := repo.db.Exec(sqlStatement, string(id))
	return err
}

func (repo *userRepository) NextID() (app.UserID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return app.UserID(id.String()), nil
}
