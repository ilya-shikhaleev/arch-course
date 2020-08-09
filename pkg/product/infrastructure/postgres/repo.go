package postgres

import (
	"database/sql"

	"github.com/ilya-shikhaleev/arch-course/pkg/product/app/product"
)

func NewProductRepository(db *sql.DB) product.Repository {
	return &repository{db: db}
}

type repository struct {
	db *sql.DB
}

func (repo *repository) FindByID(id product.ID) (*product.Product, error) {
	sqlStatement := `SELECT p.id, meta_product_id, height, color, price, title, description, material 
						FROM products AS p
						INNER JOIN meta_products AS mp ON mp.id = p.meta_product_id
						WHERE p.id=$1;`
	var p product.Product
	row := repo.db.QueryRow(sqlStatement, string(id))
	switch err := row.Scan(&p.ID, &p.MetaProductID, &p.Height, &p.Color, &p.Price, &p.Title, &p.Description, &p.Material); err {
	case sql.ErrNoRows:
		return nil, product.ErrProductNotFound
	case nil:
		return &p, nil
	default:
		return nil, err
	}
}

func (repo *repository) FindBySpecification(specification product.Specification) ([]*product.Product, error) {
	panic("implement me")
}
