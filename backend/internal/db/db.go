package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(30 * time.Minute)

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return database, nil
}

func Migrate(ctx context.Context, database *sql.DB) error {
	const query = `
CREATE TABLE IF NOT EXISTS classification_changes (
	id BIGSERIAL PRIMARY KEY,
	order_id BIGINT NOT NULL DEFAULT 1,
	position INTEGER NOT NULL,
	system_name TEXT NOT NULL,
	system_url TEXT NOT NULL DEFAULT '',
	class_before TEXT NOT NULL,
	class_after TEXT NOT NULL,
	imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_classification_changes_position ON classification_changes(position);
CREATE INDEX IF NOT EXISTS idx_classification_changes_before ON classification_changes(class_before);
CREATE INDEX IF NOT EXISTS idx_classification_changes_after ON classification_changes(class_after);

CREATE TABLE IF NOT EXISTS system_catalog (
	id BIGSERIAL PRIMARY KEY,
	order_id BIGINT NOT NULL DEFAULT 1,
	position INTEGER NOT NULL,
	code TEXT NOT NULL,
	system_name TEXT NOT NULL,
	system_url TEXT NOT NULL DEFAULT '',
	system_class TEXT NOT NULL,
	curator TEXT NOT NULL,
	imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_catalog_position ON system_catalog(position);
CREATE INDEX IF NOT EXISTS idx_system_catalog_code ON system_catalog(code);
CREATE INDEX IF NOT EXISTS idx_system_catalog_class ON system_catalog(system_class);
CREATE INDEX IF NOT EXISTS idx_system_catalog_curator ON system_catalog(curator);
`

	if _, err := database.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	const ordersQuery = `
CREATE TABLE IF NOT EXISTS orders (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO orders (id, name)
SELECT 1, 'Распоряжение 1'
WHERE NOT EXISTS (SELECT 1 FROM orders);

UPDATE orders
SET name = 'Распоряжение 1'
WHERE id = 1;

SELECT setval(pg_get_serial_sequence('orders', 'id'), GREATEST((SELECT MAX(id) FROM orders), 1), true);

ALTER TABLE classification_changes ADD COLUMN IF NOT EXISTS order_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE system_catalog ADD COLUMN IF NOT EXISTS order_id BIGINT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_classification_changes_order_id ON classification_changes(order_id);
CREATE INDEX IF NOT EXISTS idx_system_catalog_order_id ON system_catalog(order_id);
`

	if _, err := database.ExecContext(ctx, ordersQuery); err != nil {
		return fmt.Errorf("migrate orders: %w", err)
	}

	return nil
}
