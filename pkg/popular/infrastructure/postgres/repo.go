package postgres

import (
	"database/sql"

	"github.com/pkg/errors"

	"github.com/ilya-shikhaleev/arch-course/pkg/popular/app/popular"
)

func NewPopularRepository(db *sql.DB) popular.Repository {
	return &repository{db: db}
}

type repository struct {
	db *sql.DB
}

func (repo *repository) FindPopular(count int) ([]popular.Product, error) {
	sqlStatement := `SELECT product_id, title, description, material, height, color, price, buy_count						
						FROM popular 
						ORDER BY buy_count DESC LIMIT $1`
	var products []popular.Product
	rows, err := repo.db.Query(sqlStatement, count)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for rows.Next() {
		var p popular.Product
		err = rows.Scan(&p.ID, &p.Title, &p.Description, &p.Material, &p.Height, &p.Color, &p.Price, &p.BuyCount)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		products = append(products, p)
	}
	return products, errors.WithStack(err)
}

func (repo *repository) FindByID(id string) (*popular.Product, error) {
	sqlStatement := `SELECT product_id, title, description, material, height, color, price, buy_count						
						FROM popular
						WHERE product_id=$1;`
	var p popular.Product
	row := repo.db.QueryRow(sqlStatement, id)
	switch err := row.Scan(&p.ID, &p.Title, &p.Description, &p.Material, &p.Height, &p.Color, &p.Price, &p.BuyCount); err {
	case sql.ErrNoRows:
		return nil, popular.ErrProductNotFound
	case nil:
		return &p, nil
	default:
		return nil, err
	}
}

func (repo *repository) Store(p *popular.Product) error {
	sqlStatement := `
		INSERT INTO popular (product_id, title, description, material, height, color, price, buy_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (product_id) DO UPDATE SET title = EXCLUDED.title,
									   description = EXCLUDED.description,
									   material = EXCLUDED.material,
									   height = EXCLUDED.height,
									   color = EXCLUDED.color,
									   price = EXCLUDED.price,
									   buy_count = EXCLUDED.buy_count;
`
	_, err := repo.db.Exec(sqlStatement, p.ID, p.Title, p.Description, p.Material, p.Height, p.Color, p.Price, p.BuyCount)
	return err
}
