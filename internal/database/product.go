package database

import (
	"catalog-products/internal/domain"
	"catalog-products/internal/foundation/normalize"
	"context"
	"database/sql"
	"fmt"
)

// LoadAllProducts returns all products indexed by their normalized name.
func LoadAllProducts(ctx context.Context, db *sql.DB) (map[string]domain.Product, error) {
	rows, err := db.QueryContext(ctx, `SELECT Id, COALESCE(Name, ''), COALESCE(Brand, ''), COALESCE(Category, '') FROM Product`)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	products := make(map[string]domain.Product)
	for rows.Next() {
		var p domain.Product
		if err = rows.Scan(&p.Id, &p.Name, &p.Brand, &p.Category); err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products[normalize.ProductKey(p.Name, p.Brand, p.Category)] = p
	}

	return products, rows.Err()
}

// InsertProduct inserts a new product and returns its generated ID.
func InsertProduct(ctx context.Context, db Executor, p domain.Product) (int64, error) {
	var brand interface{}
	if p.Brand != "" {
		brand = p.Brand
	}

	res, err := db.ExecContext(ctx,
		`INSERT INTO Product (Name, Brand, Category) VALUES (?, ?, ?)`,
		p.Name, brand, p.Category,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert product %q: %w", p.Name, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id for product: %w", err)
	}

	return id, nil
}
