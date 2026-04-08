package database

import (
	"catalog-products/internal/domain"
	"catalog-products/internal/foundation/normalize"
	"context"
	"database/sql"
	"fmt"
)

// LoadAllSellers returns all sellers indexed by their normalized name.
func LoadAllSellers(ctx context.Context, db *sql.DB) (map[string]domain.Seller, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name FROM Seller`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sellers: %w", err)
	}
	defer rows.Close()

	sellers := make(map[string]domain.Seller)
	for rows.Next() {
		var s domain.Seller
		if err = rows.Scan(&s.Id, &s.Name); err != nil {
			return nil, fmt.Errorf("failed to scan seller row: %w", err)
		}
		sellers[normalize.String(s.Name)] = s
	}

	return sellers, rows.Err()
}

// InsertSeller inserts a new seller into the database.
func InsertSeller(ctx context.Context, db Executor, s domain.Seller) error {
	_, err := db.ExecContext(ctx, `INSERT INTO Seller (id, name) VALUES (?, ?)`, s.Id, s.Name)
	if err != nil {
		return fmt.Errorf("failed to insert seller %q: %w", s.Name, err)
	}
	return nil
}
