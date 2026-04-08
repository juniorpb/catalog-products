package catalog

import (
	"catalog-products/internal/domain"
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	schema := `
		CREATE TABLE Seller (
			id   TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		);
		CREATE TABLE Product (
			Id       INTEGER PRIMARY KEY AUTOINCREMENT,
			Name     TEXT NOT NULL,
			Brand    TEXT,
			Category TEXT
		);
		CREATE TABLE SellerProduct (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			seller_id   TEXT    NOT NULL,
			product_id  INTEGER NOT NULL,
			external_id TEXT    NOT NULL,
			FOREIGN KEY (seller_id)  REFERENCES Seller  (id),
			FOREIGN KEY (product_id) REFERENCES Product (Id),
			UNIQUE (seller_id, product_id)
		);
	`
	if _, err = db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func ptr(s string) *string { return &s }

func TestProcessEntries(t *testing.T) {
	validID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	validID2 := "b2c3d4e5-f6a7-4b5c-9d0e-1f2a3b4c5d6e"

	tests := []struct {
		name      string
		entries   []ProductEntry
		products  func() map[string]domain.Product
		sellers   func() map[string]domain.Seller
		links     func() map[string]bool
		wantStats ConsolidationStats
		wantErr   bool
	}{
		{
			name: "new product and new seller",
			entries: []ProductEntry{
				{Id: validID, SellerName: "AcmeCorp", Name: "Widget Pro", Brand: ptr("Acme"), Category: "Tools"},
			},
			products: func() map[string]domain.Product { return map[string]domain.Product{} },
			sellers:  func() map[string]domain.Seller { return map[string]domain.Seller{} },
			links:    func() map[string]bool { return map[string]bool{} },
			wantStats: ConsolidationStats{
				ProductsInserted: 1,
				LinksCreated:     1,
				LinksSkipped:     0,
			},
		},
		{
			name: "existing link is skipped",
			entries: []ProductEntry{
				{Id: validID, SellerName: "AcmeCorp", Name: "Widget Pro", Brand: ptr("Acme"), Category: "Tools"},
			},
			products: func() map[string]domain.Product {
				return map[string]domain.Product{"widget pro|acme|tools": {Id: 1, Name: "Widget Pro", Brand: "Acme", Category: "Tools"}}
			},
			sellers: func() map[string]domain.Seller {
				return map[string]domain.Seller{"acmecorp": {Id: "seller-uuid-001", Name: "AcmeCorp"}}
			},
			links: func() map[string]bool {
				return map[string]bool{"seller-uuid-001:1": true}
			},
			wantStats: ConsolidationStats{
				ProductsInserted: 0,
				LinksCreated:     0,
				LinksSkipped:     1,
			},
		},
		{
			name: "existing product linked to new seller",
			entries: []ProductEntry{
				{Id: validID2, SellerName: "NewSeller", Name: "Widget Pro", Brand: ptr("Acme"), Category: "Tools"},
			},
			products: func() map[string]domain.Product {
				return map[string]domain.Product{"widget pro|acme|tools": {Id: 1, Name: "Widget Pro", Brand: "Acme", Category: "Tools"}}
			},
			sellers:  func() map[string]domain.Seller { return map[string]domain.Seller{} },
			links:    func() map[string]bool { return map[string]bool{} },
			wantStats: ConsolidationStats{
				ProductsInserted: 0,
				LinksCreated:     1,
				LinksSkipped:     0,
			},
		},
		{
			name: "two entries same seller same product — second skipped",
			entries: []ProductEntry{
				{Id: validID, SellerName: "AcmeCorp", Name: "Widget Pro", Brand: ptr("Acme"), Category: "Tools"},
				{Id: validID2, SellerName: "AcmeCorp", Name: "Widget Pro", Brand: ptr("Acme"), Category: "Tools"},
			},
			products: func() map[string]domain.Product { return map[string]domain.Product{} },
			sellers:  func() map[string]domain.Seller { return map[string]domain.Seller{} },
			links:    func() map[string]bool { return map[string]bool{} },
			wantStats: ConsolidationStats{
				ProductsInserted: 1,
				LinksCreated:     1,
				LinksSkipped:     1,
			},
		},
		{
			name: "closed db returns error",
			entries: []ProductEntry{
				{Id: validID, SellerName: "AcmeCorp", Name: "Widget Pro", Brand: ptr("Acme"), Category: "Tools"},
			},
			products: func() map[string]domain.Product { return map[string]domain.Product{} },
			sellers:  func() map[string]domain.Seller { return map[string]domain.Seller{} },
			links:    func() map[string]bool { return map[string]bool{} },
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t)

			if tc.wantErr {
				// Close the DB to force a real database error on the first operation.
				db.Close()
			} else {
				// Pre-populate existing products into DB so FK constraints are satisfied.
				for _, p := range tc.products() {
					if p.Id > 0 {
						_, err := db.Exec(`INSERT INTO Product (Id, Name, Brand, Category) VALUES (?, ?, ?, ?)`,
							p.Id, p.Name, p.Brand, p.Category)
						if err != nil {
							t.Fatalf("setup: insert product: %v", err)
						}
					}
				}

				// Pre-populate existing sellers into DB.
				for _, s := range tc.sellers() {
					_, err := db.Exec(`INSERT INTO Seller (id, name) VALUES (?, ?)`, s.Id, s.Name)
					if err != nil {
						t.Fatalf("setup: insert seller: %v", err)
					}
				}
			}

			ctx := context.Background()
			stats, err := processEntries(ctx, db, tc.entries, tc.products(), tc.sellers(), tc.links())

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if stats != tc.wantStats {
				t.Errorf("stats mismatch\n  got:  %+v\n  want: %+v", stats, tc.wantStats)
			}
		})
	}
}
