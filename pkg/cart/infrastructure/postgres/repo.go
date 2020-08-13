package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ilya-shikhaleev/arch-course/pkg/cart/app/cart"
)

func NewCartRepository(db *sql.DB) cart.Repository {
	return &repository{db: db}
}

type repository struct {
	db *sql.DB
}

func (repo *repository) Store(cart *cart.Cart) error {
	sqlStatement := `
		INSERT INTO carts (id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET user_id  = EXCLUDED.user_id;
`
	_, err := repo.db.Exec(sqlStatement, string(cart.ID), cart.UserID)
	if err == nil {
		for _, productID := range cart.ProductIDs {
			sqlStatement := `
			INSERT INTO carts_products (cart_id, product_id)
			VALUES ($1, $2);
`
			_, err = repo.db.Exec(sqlStatement, string(cart.ID), productID)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (repo *repository) NextID() (cart.ID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return cart.ID(id.String()), nil
}

func (repo *repository) FindByID(id cart.ID) (*cart.Cart, error) {
	sqlStatement := `SELECT id, user_id
						FROM carts
						WHERE id=$1;`
	var c cart.Cart
	row := repo.db.QueryRow(sqlStatement, string(id))
	switch err := row.Scan(&c.ID, &c.UserID); err {
	case sql.ErrNoRows:
		return nil, cart.ErrCartNotFound
	case nil:
		c.ProductIDs, err = repo.findProductsByID(id)
		if err != nil {
			return nil, err
		}
		return &c, nil
	default:
		return nil, err
	}
}

func (repo *repository) FindByUserID(userID string) (*cart.Cart, error) {
	sqlStatement := `SELECT id, user_id
						FROM carts
						WHERE user_id=$1;`
	var c cart.Cart
	row := repo.db.QueryRow(sqlStatement, userID)
	switch err := row.Scan(&c.ID, &c.UserID); err {
	case sql.ErrNoRows:
		return nil, cart.ErrCartNotFound
	case nil:
		c.ProductIDs, err = repo.findProductsByID(c.ID)
		if err != nil {
			return nil, err
		}
		return &c, nil
	default:
		return nil, err
	}
}

func (repo *repository) findProductsByID(id cart.ID) ([]string, error) {
	sqlStatement := `SELECT product_id
						FROM carts_products
						WHERE cart_id=$1;`
	var products []string
	rows, err := repo.db.Query(sqlStatement, string(id))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for rows.Next() {
		var productID string
		err = rows.Scan(&productID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		products = append(products, productID)
	}
	return products, errors.WithStack(err)
}
