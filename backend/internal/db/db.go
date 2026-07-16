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

CREATE TABLE IF NOT EXISTS system_characteristics (
	id BIGSERIAL PRIMARY KEY,
	system_catalog_id BIGINT NOT NULL REFERENCES system_catalog(id) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	name TEXT NOT NULL,
	value TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_system_characteristics_catalog_id ON system_characteristics(system_catalog_id);
CREATE INDEX IF NOT EXISTS idx_system_characteristics_name ON system_characteristics(name);

CREATE TABLE IF NOT EXISTS nav_system_types (
	slug TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	position INTEGER NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nav_systems (
	system_key TEXT PRIMARY KEY,
	system_name TEXT NOT NULL,
	system_url TEXT NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nav_system_characteristics (
	system_key TEXT NOT NULL REFERENCES nav_systems(system_key) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	name TEXT NOT NULL,
	value TEXT NOT NULL,
	PRIMARY KEY (system_key, position)
);

CREATE INDEX IF NOT EXISTS idx_nav_system_characteristics_name ON nav_system_characteristics(name);

CREATE TABLE IF NOT EXISTS nav_parser_settings (
	id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
	update_interval_days INTEGER NOT NULL DEFAULT 7 CHECK (update_interval_days BETWEEN 1 AND 365),
	last_run_at TIMESTAMPTZ
);

INSERT INTO nav_parser_settings (id, update_interval_days)
VALUES (TRUE, 7)
ON CONFLICT (id) DO NOTHING;
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
	ALTER TABLE classification_changes ADD COLUMN IF NOT EXISTS construction_type TEXT NOT NULL DEFAULT 'Тип не присвоен';
	ALTER TABLE system_catalog ADD COLUMN IF NOT EXISTS order_id BIGINT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_classification_changes_order_id ON classification_changes(order_id);
CREATE INDEX IF NOT EXISTS idx_system_catalog_order_id ON system_catalog(order_id);

CREATE TABLE IF NOT EXISTS system_documents (
	id BIGSERIAL PRIMARY KEY,
	order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
	system_catalog_id BIGINT NOT NULL REFERENCES system_catalog(id) ON DELETE CASCADE,
	comment TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (system_catalog_id)
);

CREATE INDEX IF NOT EXISTS idx_system_documents_order_id ON system_documents(order_id);
CREATE INDEX IF NOT EXISTS idx_system_documents_catalog_id ON system_documents(system_catalog_id);

ALTER TABLE system_documents ADD COLUMN IF NOT EXISTS comparison_selected BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE system_documents ADD COLUMN IF NOT EXISTS attachment_name TEXT NOT NULL DEFAULT '';
ALTER TABLE system_documents ADD COLUMN IF NOT EXISTS attachment_content_type TEXT NOT NULL DEFAULT '';
ALTER TABLE system_documents ADD COLUMN IF NOT EXISTS attachment_size BIGINT NOT NULL DEFAULT 0;
ALTER TABLE system_documents ADD COLUMN IF NOT EXISTS attachment_data BYTEA;
ALTER TABLE system_catalog ADD COLUMN IF NOT EXISTS document_initialized BOOLEAN NOT NULL DEFAULT FALSE;

INSERT INTO system_documents (order_id, system_catalog_id)
SELECT s.order_id, s.id
FROM system_catalog s
WHERE NOT s.document_initialized AND NOT EXISTS (
	SELECT 1 FROM system_documents d WHERE d.system_catalog_id = s.id
);

UPDATE system_catalog SET document_initialized = TRUE WHERE NOT document_initialized;

WITH ranked_sources AS (
	SELECT
		s.id,
		LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g')) AS system_key,
		s.system_name,
		s.system_url,
		ROW_NUMBER() OVER (
			PARTITION BY LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g'))
			ORDER BY (SELECT COUNT(*) FROM system_characteristics c WHERE c.system_catalog_id = s.id) DESC,
				s.imported_at DESC,
				s.id DESC
		) AS source_rank
	FROM system_catalog s
	WHERE NULLIF(BTRIM(s.system_url), '') IS NOT NULL
		OR EXISTS (SELECT 1 FROM system_characteristics c WHERE c.system_catalog_id = s.id)
), metadata_sources AS (
	SELECT id, system_key, system_name, system_url
	FROM ranked_sources
	WHERE source_rank = 1
)
INSERT INTO nav_systems (system_key, system_name, system_url)
SELECT system_key, system_name, system_url
FROM metadata_sources
ON CONFLICT (system_key) DO NOTHING;

WITH ranked_sources AS (
	SELECT
		s.id,
		LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g')) AS system_key,
		ROW_NUMBER() OVER (
			PARTITION BY LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g'))
			ORDER BY (SELECT COUNT(*) FROM system_characteristics c WHERE c.system_catalog_id = s.id) DESC,
				s.imported_at DESC,
				s.id DESC
		) AS source_rank
	FROM system_catalog s
	WHERE EXISTS (SELECT 1 FROM system_characteristics c WHERE c.system_catalog_id = s.id)
), metadata_sources AS (
	SELECT id, system_key
	FROM ranked_sources
	WHERE source_rank = 1
)
INSERT INTO nav_system_characteristics (system_key, position, name, value)
SELECT source.system_key, characteristic.position, characteristic.name, characteristic.value
FROM metadata_sources source
JOIN system_characteristics characteristic ON characteristic.system_catalog_id = source.id
WHERE NOT EXISTS (
	SELECT 1
	FROM nav_system_characteristics existing
	WHERE existing.system_key = source.system_key
)
ON CONFLICT (system_key, position) DO NOTHING;
`

	if _, err := database.ExecContext(ctx, ordersQuery); err != nil {
		return fmt.Errorf("migrate orders: %w", err)
	}

	return nil
}
