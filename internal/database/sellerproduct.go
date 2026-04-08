package database

import (
	"catalog-products/internal/domain"
	"context"
	"database/sql"
	"fmt"
)

// LoadAllSellerProducts returns a set of all existing seller-product links.
// The map key is "sellerID:productID".
func LoadAllSellerProducts(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT seller_id, product_id FROM SellerProduct`)
	if err != nil {
		return nil, fmt.Errorf("failed to query seller products: %w", err)
	}
	defer rows.Close()

	links := make(map[string]bool)
	for rows.Next() {
		var sellerID string
		var productID int64
		if err = rows.Scan(&sellerID, &productID); err != nil {
			return nil, fmt.Errorf("failed to scan seller product row: %w", err)
		}
		links[sellerProductKey(sellerID, productID)] = true
	}

	return links, rows.Err()
}

// InsertSellerProduct inserts a new seller-product link.
func InsertSellerProduct(ctx context.Context, db Executor, sp domain.SellerProduct) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO SellerProduct (seller_id, product_id, external_id) VALUES (?, ?, ?)`,
		sp.SellerId, sp.ProductId, sp.ExternalId,
	)
	if err != nil {
		return fmt.Errorf("failed to insert seller product (seller=%s, product=%d): %w", sp.SellerId, sp.ProductId, err)
	}
	return nil
}

func sellerProductKey(sellerID string, productID int64) string {
	return fmt.Sprintf("%s:%d", sellerID, productID)
}
