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
	database.SetConnMaxIdleTime(5 * time.Minute)
	database.SetConnMaxLifetime(30 * time.Minute)

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return database, nil
}

func Migrate(ctx context.Context, database *sql.DB) error {
	if err := ensureMigrationsTable(ctx, database); err != nil {
		return err
	}

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
		image_url TEXT NOT NULL DEFAULT '',
		image_content_type TEXT NOT NULL DEFAULT '',
		image_data BYTEA,
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
		update_interval_days INTEGER NOT NULL DEFAULT 1 CHECK (update_interval_days BETWEEN 1 AND 365),
		worker_count INTEGER NOT NULL DEFAULT 4 CHECK (worker_count BETWEEN 1 AND 10),
		request_timeout_seconds INTEGER NOT NULL DEFAULT 35 CHECK (request_timeout_seconds BETWEEN 5 AND 120),
		retry_attempts INTEGER NOT NULL DEFAULT 3 CHECK (retry_attempts BETWEEN 1 AND 5),
		retry_delay_seconds INTEGER NOT NULL DEFAULT 2 CHECK (retry_delay_seconds BETWEEN 1 AND 30),
		fallback_search BOOLEAN NOT NULL DEFAULT TRUE,
		last_run_at TIMESTAMPTZ
	);

	CREATE TABLE IF NOT EXISTS nav_parser_runs (
		id BIGSERIAL PRIMARY KEY,
		source TEXT NOT NULL DEFAULT 'manual',
		status TEXT NOT NULL,
		message TEXT NOT NULL DEFAULT '',
		total INTEGER NOT NULL DEFAULT 0,
		found INTEGER NOT NULL DEFAULT 0,
		updated INTEGER NOT NULL DEFAULT 0,
		failed INTEGER NOT NULL DEFAULT 0,
		not_found INTEGER NOT NULL DEFAULT 0,
		started_at TIMESTAMPTZ NOT NULL,
		finished_at TIMESTAMPTZ NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_nav_parser_runs_started_at ON nav_parser_runs(started_at DESC);

	CREATE TABLE IF NOT EXISTS nav_parser_run_logs (
		id BIGSERIAL PRIMARY KEY,
		run_id BIGINT NOT NULL REFERENCES nav_parser_runs(id) ON DELETE CASCADE,
		logged_at TIMESTAMPTZ NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_nav_parser_run_logs_run_id ON nav_parser_run_logs(run_id, id);

	ALTER TABLE nav_parser_settings ADD COLUMN IF NOT EXISTS worker_count INTEGER NOT NULL DEFAULT 4;
	ALTER TABLE nav_parser_settings ADD COLUMN IF NOT EXISTS request_timeout_seconds INTEGER NOT NULL DEFAULT 35;
	ALTER TABLE nav_parser_settings ADD COLUMN IF NOT EXISTS retry_attempts INTEGER NOT NULL DEFAULT 3;
	ALTER TABLE nav_parser_settings ADD COLUMN IF NOT EXISTS retry_delay_seconds INTEGER NOT NULL DEFAULT 2;
	ALTER TABLE nav_parser_settings ADD COLUMN IF NOT EXISTS fallback_search BOOLEAN NOT NULL DEFAULT TRUE;
	ALTER TABLE nav_parser_settings ALTER COLUMN update_interval_days SET DEFAULT 1;

	INSERT INTO nav_parser_settings (id, update_interval_days)
	VALUES (TRUE, 1)
	ON CONFLICT (id) DO NOTHING;
	`

	if err := applyMigration(ctx, database, 1, query); err != nil {
		return fmt.Errorf("apply base schema migration: %w", err)
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

	SELECT setval(pg_get_serial_sequence('orders', 'id'), GREATEST((SELECT MAX(id) FROM orders), 1), true);

	ALTER TABLE classification_changes ADD COLUMN IF NOT EXISTS order_id BIGINT NOT NULL DEFAULT 1;
	ALTER TABLE classification_changes ADD COLUMN IF NOT EXISTS construction_type TEXT NOT NULL DEFAULT 'Тип не присвоен';
		ALTER TABLE system_catalog ADD COLUMN IF NOT EXISTS order_id BIGINT NOT NULL DEFAULT 1;
		ALTER TABLE nav_system_types ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';
	ALTER TABLE nav_system_types ADD COLUMN IF NOT EXISTS image_content_type TEXT NOT NULL DEFAULT '';
	ALTER TABLE nav_system_types ADD COLUMN IF NOT EXISTS image_data BYTEA;

	CREATE INDEX IF NOT EXISTS idx_classification_changes_order_id ON classification_changes(order_id);
	CREATE INDEX IF NOT EXISTS idx_system_catalog_order_id ON system_catalog(order_id);
	CREATE INDEX IF NOT EXISTS idx_classification_changes_order_position ON classification_changes(order_id, position, id);
	CREATE INDEX IF NOT EXISTS idx_classification_changes_order_after ON classification_changes(order_id, class_after);
	CREATE INDEX IF NOT EXISTS idx_system_catalog_order_position ON system_catalog(order_id, position, id);
	CREATE INDEX IF NOT EXISTS idx_system_catalog_order_class ON system_catalog(order_id, system_class);
	CREATE INDEX IF NOT EXISTS idx_system_catalog_order_curator ON system_catalog(order_id, curator);

	DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conname = 'classification_changes_order_id_fkey'
				AND conrelid = 'classification_changes'::regclass
		) THEN
			ALTER TABLE classification_changes
				ADD CONSTRAINT classification_changes_order_id_fkey
				FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE NOT VALID;
		END IF;
		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conname = 'system_catalog_order_id_fkey'
				AND conrelid = 'system_catalog'::regclass
		) THEN
			ALTER TABLE system_catalog
				ADD CONSTRAINT system_catalog_order_id_fkey
				FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE NOT VALID;
		END IF;
	END $$;

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

	UPDATE classification_changes change
SET construction_type = characteristic.value
FROM nav_systems nav
JOIN nav_system_characteristics characteristic
	ON characteristic.system_key = nav.system_key
	AND LOWER(BTRIM(characteristic.name)) IN ('сегмент строительства', 'тип строительства')
WHERE change.construction_type = 'Тип не присвоен'
	AND nav.system_key = LOWER(REGEXP_REPLACE(BTRIM(change.system_name), '\s+', ' ', 'g'))
	AND characteristic.value IN (
		'Промышленное и гражданское строительство',
		'Индивидуальное жилищное строительство',
		'Транспортное и дорожное строительство',
		'Специальные сооружения'
	);

UPDATE classification_changes change
SET construction_type = known.construction_type,
	system_url = CASE WHEN known.system_url <> '' THEN known.system_url ELSE change.system_url END
FROM (VALUES
	('тн-гео полигон фрост', 'https://nav.tn.ru/systems/poligony-ploshchadki-khraneniya-i-pr/tn-geo-poligon-frost/', 'Специальные сооружения'),
	('тн-гео хвостохранилище фрост', 'https://nav.tn.ru/systems/iskusstvennye-vodoemy-prudy-i-pr/tn-geo-khvostokhranilishche-frost/', 'Специальные сооружения'),
	('тн-гео амбар шламовый фрост', 'https://nav.tn.ru/systems/iskusstvennye-vodoemy-prudy-i-pr/tn-geo-ambar-shlamovyy-frost/', 'Специальные сооружения'),
	('тн-авиа впп фрост', 'https://nav.tn.ru/systems/konstruktsiya-letnogo-polya/tn-avia-vpp-frost/', 'Транспортное и дорожное строительство'),
	('тн-кровля солид керамзит', 'https://nav.tn.ru/systems/ploskaya-krysha/tn-krovlya-solid-keramzit/', 'Промышленное и гражданское строительство'),
	('тн-техизоляция камин', '', 'Индивидуальное жилищное строительство')
) AS known(system_key, system_url, construction_type)
WHERE change.construction_type = 'Тип не присвоен'
	AND REGEXP_REPLACE(
		LOWER(REGEXP_REPLACE(BTRIM(change.system_name), '\s+', ' ', 'g')),
		'^система\s+',
		''
		) = known.system_key;

	ALTER TABLE classification_changes VALIDATE CONSTRAINT classification_changes_order_id_fkey;
	ALTER TABLE system_catalog VALIDATE CONSTRAINT system_catalog_order_id_fkey;
	`

	if err := applyMigration(ctx, database, 2, ordersQuery); err != nil {
		return fmt.Errorf("apply orders schema migration: %w", err)
	}

	const constraintsQuery = `
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'classification_changes_class_before_check' AND conrelid = 'classification_changes'::regclass) THEN
			ALTER TABLE classification_changes ADD CONSTRAINT classification_changes_class_before_check
				CHECK (class_before IN ('Новая система', 'Рекомендованная', 'Разрешенная', 'Запрещенная')) NOT VALID;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'classification_changes_class_after_check' AND conrelid = 'classification_changes'::regclass) THEN
			ALTER TABLE classification_changes ADD CONSTRAINT classification_changes_class_after_check
				CHECK (class_after IN ('Рекомендованная', 'Разрешенная', 'Запрещенная')) NOT VALID;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'system_catalog_class_check' AND conrelid = 'system_catalog'::regclass) THEN
			ALTER TABLE system_catalog ADD CONSTRAINT system_catalog_class_check
				CHECK (system_class IN ('Рекомендованная', 'Разрешенная', 'Запрещенная')) NOT VALID;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'nav_parser_runs_status_check' AND conrelid = 'nav_parser_runs'::regclass) THEN
			ALTER TABLE nav_parser_runs ADD CONSTRAINT nav_parser_runs_status_check
				CHECK (status IN ('completed', 'failed', 'canceled')) NOT VALID;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'nav_parser_runs_source_check' AND conrelid = 'nav_parser_runs'::regclass) THEN
			ALTER TABLE nav_parser_runs ADD CONSTRAINT nav_parser_runs_source_check
				CHECK (source IN ('manual', 'scheduled')) NOT VALID;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'nav_parser_run_logs_level_check' AND conrelid = 'nav_parser_run_logs'::regclass) THEN
			ALTER TABLE nav_parser_run_logs ADD CONSTRAINT nav_parser_run_logs_level_check
				CHECK (level IN ('info', 'success', 'warning', 'error')) NOT VALID;
		END IF;
	END $$;
	`
	if err := applyMigration(ctx, database, 3, constraintsQuery); err != nil {
		return fmt.Errorf("apply domain constraints migration: %w", err)
	}

	const parserScheduleQuery = `
	ALTER TABLE nav_parser_settings ADD COLUMN IF NOT EXISTS last_attempt_at TIMESTAMPTZ;
	ALTER TABLE nav_parser_settings ADD COLUMN IF NOT EXISTS consecutive_failures INTEGER NOT NULL DEFAULT 0;
	`
	if err := applyMigration(ctx, database, 4, parserScheduleQuery); err != nil {
		return fmt.Errorf("apply parser schedule migration: %w", err)
	}

	const queryIndexes = `
	CREATE INDEX IF NOT EXISTS idx_classification_changes_order_system_name
		ON classification_changes (order_id, (LOWER(BTRIM(system_name))));
	CREATE INDEX IF NOT EXISTS idx_system_catalog_order_code
		ON system_catalog (order_id, code);
	CREATE INDEX IF NOT EXISTS idx_system_documents_order_comparison
		ON system_documents (order_id, comparison_selected);
	`
	if err := applyMigration(ctx, database, 5, queryIndexes); err != nil {
		return fmt.Errorf("apply query indexes migration: %w", err)
	}

	const normalizeConstructionTypesQuery = `
	UPDATE nav_system_characteristics
	SET value = CASE
		WHEN LOWER(value) = 'пгс' OR LOWER(value) LIKE '%промышлен%' OR LOWER(value) LIKE '%гражданск%'
			THEN 'Промышленное и гражданское строительство'
		WHEN LOWER(value) = 'ижс' OR LOWER(value) LIKE '%индивидуальн%' OR LOWER(value) LIKE '%жилищн%'
			THEN 'Индивидуальное жилищное строительство'
		WHEN LOWER(value) LIKE '%транспорт%' OR LOWER(value) LIKE '%дорож%'
			THEN 'Транспортное и дорожное строительство'
		WHEN LOWER(value) = 'сс' OR LOWER(value) LIKE '%специальн%' OR LOWER(value) LIKE '%спецсооруж%' OR LOWER(value) LIKE '%спец сооруж%'
			THEN 'Специальные сооружения'
		ELSE value
	END
	WHERE LOWER(BTRIM(name)) IN ('сегмент строительства', 'тип строительства');
	`
	if err := applyMigration(ctx, database, 6, normalizeConstructionTypesQuery); err != nil {
		return fmt.Errorf("normalize parsed construction types migration: %w", err)
	}

	const removeUnsafeImagesQuery = `
	UPDATE nav_system_types
	SET image_content_type = '', image_data = NULL
	WHERE image_data IS NOT NULL
		AND image_content_type NOT IN ('image/png', 'image/jpeg', 'image/gif', 'image/webp');
	`
	if err := applyMigration(ctx, database, 7, removeUnsafeImagesQuery); err != nil {
		return fmt.Errorf("remove unsafe parsed images migration: %w", err)
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, database *sql.DB) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migrations table setup: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(914_070_221)); err != nil {
		return fmt.Errorf("lock migrations table setup: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations table setup: %w", err)
	}
	return nil
}

func applyMigration(ctx context.Context, database *sql.DB, version int64, query string) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", version, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(914_070_221)); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	var applied bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)
	`, version).Scan(&applied); err != nil {
		return fmt.Errorf("check migration %d: %w", version, err)
	}
	if applied {
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("execute migration %d: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("record migration %d: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", version, err)
	}
	return nil
}
