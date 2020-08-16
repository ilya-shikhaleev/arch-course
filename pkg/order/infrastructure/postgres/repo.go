package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ilya-shikhaleev/arch-course/pkg/order/app/order"
)

func NewOrderRepository(db *sql.DB) order.Repository {
	return &repository{db: db}
}

type repository struct {
	db *sql.DB
}

func (repo *repository) Store(order *order.Order) error {
	sqlStatement := `
		INSERT INTO orders (id, user_id, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET user_id = EXCLUDED.user_id, status = EXCLUDED.status;
`
	_, err := repo.db.Exec(sqlStatement, string(order.ID), order.UserID, order.Status)
	if err == nil {
		for _, product := range order.Products {
			sqlStatement := `
			INSERT INTO orders_products (order_id, product_id, price)
			VALUES ($1, $2, $3);
`
			_, err = repo.db.Exec(sqlStatement, string(order.ID), product.ProductID, product.Price)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (repo *repository) NextID() (order.ID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return order.ID(id.String()), nil
}

func (repo *repository) FindByID(id order.ID) (*order.Order, error) {
	sqlStatement := `SELECT id, user_id, status
						FROM orders
						WHERE id=$1;`
	var o order.Order
	row := repo.db.QueryRow(sqlStatement, string(id))
	switch err := row.Scan(&o.ID, &o.UserID, &o.Status); err {
	case sql.ErrNoRows:
		return nil, order.ErrOrderNotFound
	case nil:
		o.Products, err = repo.findProductsByID(id)
		if err != nil {
			return nil, err
		}
		return &o, nil
	default:
		return nil, err
	}
}

func (repo *repository) FindByUserID(userID string) ([]order.Order, error) {
	sqlStatement := `SELECT id, user_id, status
						FROM orders
						WHERE user_id=$1;`
	var orders []order.Order
	rows, err := repo.db.Query(sqlStatement, userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for rows.Next() {
		var o order.Order
		err = rows.Scan(&o.ID, &o.UserID, &o.Status)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		o.Products, err = repo.findProductsByID(o.ID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		orders = append(orders, o)
	}
	return orders, errors.WithStack(err)
}

func (repo *repository) findProductsByID(id order.ID) ([]order.Product, error) {
	sqlStatement := `SELECT product_id, price
						FROM orders_products
						WHERE order_id=$1;`
	var products []order.Product
	rows, err := repo.db.Query(sqlStatement, string(id))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for rows.Next() {
		var product order.Product
		err = rows.Scan(&product.ProductID, &product.Price)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		products = append(products, product)
	}
	return products, errors.WithStack(err)
}
