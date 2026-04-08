package main

import (
	"catalog-products/internal/business/catalog"
	"catalog-products/internal/database"
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	if err := database.ConnectDB(ctx); err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	if err := database.RunMigrations(ctx); err != nil {
		log.Fatal(err)
	}

	if err := catalog.Consolidate(ctx, database.DB); err != nil {
		log.Fatal(err)
	}
}
