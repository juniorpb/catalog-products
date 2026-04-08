package catalog

import (
	"catalog-products/internal/database"
	"catalog-products/internal/domain"
	"catalog-products/internal/foundation/normalize"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ConsolidationStats holds counters for a single consolidation run.
type ConsolidationStats struct {
	ProductsInserted int
	LinksCreated     int
	LinksSkipped     int
}

// Consolidate reads the seller catalog file, normalizes the entries, and
// persists new products and seller links into the database without duplicating
// existing records. All writes are wrapped in a single transaction.
func Consolidate(ctx context.Context, db *sql.DB) error {
	jsonPath := filepath.Join(projectRoot(), "data", "ProductEntry.json")

	entries, err := ParseJSONFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read catalog: %w", err)
	}

	entries = sanitizeEntries(entries)
	entries = deduplicateByExternalID(entries)

	products, err := database.LoadAllProducts(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to load products: %w", err)
	}

	sellers, err := database.LoadAllSellers(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to load sellers: %w", err)
	}

	links, err := database.LoadAllSellerProducts(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to load seller products: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stats, err := processEntries(ctx, tx, entries, products, sellers, links)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("consolidation done — products inserted: %d | links created: %d | skipped: %d\n",
		stats.ProductsInserted, stats.LinksCreated, stats.LinksSkipped)

	return nil
}

// processEntries iterates over sanitized entries and persists new sellers,
// products and seller-product links. Maps are mutated in place to avoid
// redundant DB round-trips within the same run.
func processEntries(
	ctx context.Context,
	db database.Executor,
	entries []ProductEntry,
	products map[string]domain.Product,
	sellers map[string]domain.Seller,
	links map[string]bool,
) (ConsolidationStats, error) {
	var stats ConsolidationStats

	for _, e := range entries {
		seller, err := ensureSeller(ctx, db, sellers, e.SellerName)
		if err != nil {
			return ConsolidationStats{}, err
		}

		brandStr := ""
		if e.Brand != nil {
			brandStr = *e.Brand
		}
		key := normalize.ProductKey(e.Name, brandStr, e.Category)
		isNewProduct := products[key].Id == 0

		product, err := ensureProduct(ctx, db, products, e)
		if err != nil {
			return ConsolidationStats{}, err
		}

		if isNewProduct {
			stats.ProductsInserted++
		}

		linkKey := fmt.Sprintf("%s:%d", seller.Id, product.Id)
		if links[linkKey] {
			stats.LinksSkipped++
			continue
		}

		sp := domain.SellerProduct{
			SellerId:   seller.Id,
			ProductId:  product.Id,
			ExternalId: e.Id,
		}
		if err = database.InsertSellerProduct(ctx, db, sp); err != nil {
			return ConsolidationStats{}, err
		}

		links[linkKey] = true
		stats.LinksCreated++
	}

	return stats, nil
}

// ensureSeller returns the seller from the map or inserts and registers it.
func ensureSeller(ctx context.Context, db database.Executor, sellers map[string]domain.Seller, name string) (domain.Seller, error) {
	key := normalize.String(name)
	if s, ok := sellers[key]; ok {
		return s, nil
	}

	s := domain.Seller{Id: uuid.New().String(), Name: name}
	if err := database.InsertSeller(ctx, db, s); err != nil {
		return domain.Seller{}, err
	}

	sellers[key] = s
	return s, nil
}

// ensureProduct returns the product from the map or inserts and registers it.
func ensureProduct(ctx context.Context, db database.Executor, products map[string]domain.Product, e ProductEntry) (domain.Product, error) {
	brand := ""
	if e.Brand != nil {
		brand = *e.Brand
	}

	key := normalize.ProductKey(e.Name, brand, e.Category)
	if p, ok := products[key]; ok {
		return p, nil
	}

	p := domain.Product{Name: e.Name, Brand: brand, Category: e.Category}
	id, err := database.InsertProduct(ctx, db, p)
	if err != nil {
		return domain.Product{}, err
	}

	p.Id = id
	products[key] = p
	return p, nil
}

// sanitizeEntries normalizes names and replaces invalid UUIDs with new ones.
func sanitizeEntries(entries []ProductEntry) []ProductEntry {
	result := make([]ProductEntry, 0, len(entries))
	for _, e := range entries {
		e.Name = strings.TrimSpace(e.Name)
		e.SellerName = strings.TrimSpace(e.SellerName)
		e.Category = strings.TrimSpace(e.Category)

		if e.Name == "" || e.SellerName == "" {
			continue
		}

		if !normalize.IsValidUUID(e.Id) {
			e.Id = uuid.New().String()
		}

		result = append(result, e)
	}
	return result
}

// deduplicateByExternalID removes duplicate entries sharing the same external ID,
// keeping the first occurrence.
func deduplicateByExternalID(entries []ProductEntry) []ProductEntry {
	seen := make(map[string]bool, len(entries))
	result := make([]ProductEntry, 0, len(entries))
	for _, e := range entries {
		if seen[e.Id] {
			continue
		}

		seen[e.Id] = true
		result = append(result, e)
	}

	return result
}

func projectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}
	return cwd
}
