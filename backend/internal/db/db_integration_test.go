package db

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMigrateIsRepeatableAndDoesNotRenameOrders(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	defer database.Close()

	migrationErrors := make(chan error, 2)
	for range 2 {
		go func() { migrationErrors <- Migrate(ctx, database) }()
	}
	for range 2 {
		if err := <-migrationErrors; err != nil {
			t.Fatalf("concurrent migration: %v", err)
		}
	}
	if _, err := database.ExecContext(ctx, `UPDATE orders SET name = 'Собственное название' WHERE id = 1`); err != nil {
		t.Fatalf("rename test order: %v", err)
	}
	if err := Migrate(ctx, database); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}

	var name string
	if err := database.QueryRowContext(ctx, `SELECT name FROM orders WHERE id = 1`).Scan(&name); err != nil {
		t.Fatalf("load test order: %v", err)
	}
	if name != "Собственное название" {
		t.Fatalf("migration unexpectedly renamed the order to %q", name)
	}
}
