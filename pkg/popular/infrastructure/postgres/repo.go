package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"

	"github.com/ilya-shikhaleev/arch-course/pkg/popular/app/popular"
)

func NewPopularRepository(db *sql.DB, client *redis.Client) popular.Repository {
	return &repository{db: db, client: client}
}

const redisKey = "products"

type cache struct {
	Products []popular.Product `json:"products,omitempty"`
}

type repository struct {
	db     *sql.DB
	client *redis.Client
}

func (repo *repository) FindPopular(count int) ([]popular.Product, error) {
	products, err := repo.readFromCache()
	if err == nil {
		return products, nil
	}

	sqlStatement := `SELECT product_id, title, description, material, height, color, price, buy_count						
						FROM popular 
						ORDER BY buy_count DESC LIMIT $1`
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
	_ = repo.writeToCache(products)

	return products, errors.WithStack(err)
}

func (repo *repository) readFromCache() ([]popular.Product, error) {
	cacheStr, err := repo.client.Get(redisKey).Result()
	if err != nil {
		return nil, err
	}

	var cache cache
	err = json.Unmarshal([]byte(cacheStr), &cache)
	if err != nil {
		return nil, err
	}

	return cache.Products, nil
}

func (repo *repository) writeToCache(products []popular.Product) error {
	if len(products) == 0 {
		return nil
	}
	cache := cache{Products: products}
	cacheBytes, err := json.Marshal(&cache)
	if err != nil {
		return err
	}
	return repo.client.Set(redisKey, string(cacheBytes), 10*time.Second).Err()
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
