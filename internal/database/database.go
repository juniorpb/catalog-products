package database

import (
	"catalog-products/internal/foundation/files"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func projectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}
	return cwd
}

func ConnectDB(ctx context.Context) error {
	dbPath := filepath.Join(projectRoot(), "data", "catalog.db")

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("database connected:", dbPath)
	return nil
}

func RunMigrations(ctx context.Context) error {
	migrationsDir := filepath.Join(projectRoot(), "internal", "database", "migrations")

	sqls, err := files.ReadSQLFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, query := range sqls {
		if _, err = DB.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	log.Printf("%d migration(s) executed\n", len(sqls))
	return nil
}
