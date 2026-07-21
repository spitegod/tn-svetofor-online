package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tn/backend/internal/model"
)

type SystemCatalogRepository struct {
	db *sql.DB
}

func NewSystemCatalogRepository(db *sql.DB) *SystemCatalogRepository {
	return &SystemCatalogRepository{db: db}
}

func (r *SystemCatalogRepository) ReplaceAll(ctx context.Context, orderID int64, rows []model.SystemCatalogRow) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace system catalog: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	type documentState struct {
		comment            string
		comparisonSelected bool
		attachmentName     string
		attachmentType     string
		attachmentSize     int64
		attachmentData     []byte
	}
	documentStates := make(map[string]documentState)
	existing, queryErr := tx.QueryContext(ctx, `
		SELECT s.code, s.system_name, d.comment, d.comparison_selected,
			d.attachment_name, d.attachment_content_type, d.attachment_size, d.attachment_data
		FROM system_documents d
		JOIN system_catalog s ON s.id = d.system_catalog_id
		WHERE d.order_id = $1
	`, orderID)
	if queryErr != nil {
		return fmt.Errorf("load system document comments before catalog import: %w", queryErr)
	}
	for existing.Next() {
		var code, name string
		var state documentState
		if err = existing.Scan(&code, &name, &state.comment, &state.comparisonSelected,
			&state.attachmentName, &state.attachmentType, &state.attachmentSize, &state.attachmentData); err != nil {
			existing.Close()
			return fmt.Errorf("scan preserved system document comment: %w", err)
		}
		documentStates[code+"\x00"+name] = state
	}
	if err = existing.Close(); err != nil {
		return fmt.Errorf("close preserved system document comments: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM system_catalog WHERE order_id = $1`, orderID); err != nil {
		return fmt.Errorf("clear system catalog: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO system_catalog (order_id, position, code, system_name, system_url, system_class, curator, document_initialized)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING id
	`)
	if err != nil {
		return fmt.Errorf("prepare system catalog insert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		state := documentStates[row.Code+"\x00"+row.SystemName]
		var systemCatalogID int64
		if err = stmt.QueryRowContext(ctx, orderID, row.Position, row.Code, row.SystemName, row.SystemURL, row.SystemClass, row.Curator).Scan(&systemCatalogID); err != nil {
			return fmt.Errorf("insert system catalog row %q: %w", row.Code, err)
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO system_documents (
				order_id, system_catalog_id, comment, comparison_selected,
				attachment_name, attachment_content_type, attachment_size, attachment_data
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, orderID, systemCatalogID, state.comment, state.comparisonSelected,
			state.attachmentName, state.attachmentType, state.attachmentSize, state.attachmentData); err != nil {
			return fmt.Errorf("insert system document row %q: %w", row.Code, err)
		}
	}

	if _, err = tx.ExecContext(ctx, `UPDATE orders SET updated_at = NOW() WHERE id = $1`, orderID); err != nil {
		return fmt.Errorf("touch order after system catalog import: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit system catalog: %w", err)
	}

	return nil
}

func (r *SystemCatalogRepository) Update(ctx context.Context, id int64, orderID int64, row model.SystemCatalogRow) (model.SystemCatalogRow, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SystemCatalogRow{}, fmt.Errorf("begin system catalog update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var updated model.SystemCatalogRow
	err = tx.QueryRowContext(ctx, `
		UPDATE system_catalog
		SET code = $3, system_name = $4, system_class = $5, curator = $6
		WHERE id = $1 AND order_id = $2
		RETURNING id, order_id, position, code, system_name, system_url, system_class, curator, imported_at
	`, id, orderID, row.Code, row.SystemName, row.SystemClass, row.Curator).Scan(
		&updated.ID, &updated.OrderID, &updated.Position, &updated.Code, &updated.SystemName,
		&updated.SystemURL, &updated.SystemClass, &updated.Curator, &updated.ImportedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.SystemCatalogRow{}, fmt.Errorf("system catalog row not found")
		}
		return model.SystemCatalogRow{}, fmt.Errorf("update system catalog row: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `UPDATE orders SET updated_at = NOW() WHERE id = $1`, orderID); err != nil {
		return model.SystemCatalogRow{}, fmt.Errorf("touch order after system catalog update: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return model.SystemCatalogRow{}, fmt.Errorf("commit system catalog update: %w", err)
	}
	return updated, nil
}

func (r *SystemCatalogRepository) List(ctx context.Context, filter model.SystemCatalogFilter) ([]model.SystemCatalogRow, error) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if filter.OrderID > 0 {
		args = append(args, filter.OrderID)
		clauses = append(clauses, fmt.Sprintf("s.order_id = $%d", len(args)))
	}

	if filter.Query != "" {
		args = append(args, "%"+strings.ToLower(filter.Query)+"%")
		clauses = append(clauses, fmt.Sprintf("(LOWER(s.system_name) LIKE $%d OR LOWER(s.code) LIKE $%d)", len(args), len(args)))
	}

	if filter.SystemClass != "" {
		args = append(args, filter.SystemClass)
		clauses = append(clauses, fmt.Sprintf("s.system_class = $%d", len(args)))
	}

	if filter.Curator != "" {
		args = append(args, filter.Curator)
		clauses = append(clauses, fmt.Sprintf("s.curator = $%d", len(args)))
	}

	query := `
		SELECT s.id, s.order_id, s.position, s.code, s.system_name,
			COALESCE(NULLIF(nav.system_url, ''), s.system_url),
			s.system_class, s.curator, s.imported_at
		FROM system_catalog s
		LEFT JOIN nav_systems nav
			ON nav.system_key = LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g'))
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY s.position, s.id"

	result, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list system catalog: %w", err)
	}
	defer result.Close()

	rows := make([]model.SystemCatalogRow, 0)
	for result.Next() {
		var row model.SystemCatalogRow
		if err := result.Scan(&row.ID, &row.OrderID, &row.Position, &row.Code, &row.SystemName, &row.SystemURL, &row.SystemClass, &row.Curator, &row.ImportedAt); err != nil {
			return nil, fmt.Errorf("scan system catalog row: %w", err)
		}
		rows = append(rows, row)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate system catalog: %w", err)
	}

	if err := r.loadCharacteristics(ctx, rows); err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *SystemCatalogRepository) ParserRows(ctx context.Context) ([]model.SystemCatalogRow, error) {
	result, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, position, code, system_name, system_url, system_class, curator, imported_at
		FROM (
			SELECT DISTINCT ON (LOWER(REGEXP_REPLACE(BTRIM(system_name), '\s+', ' ', 'g')))
				id, order_id, position, code, system_name, system_url, system_class, curator, imported_at
			FROM system_catalog
			ORDER BY LOWER(REGEXP_REPLACE(BTRIM(system_name), '\s+', ' ', 'g')), imported_at DESC, id DESC
		) unique_systems
		ORDER BY LOWER(system_name), id
	`)
	if err != nil {
		return nil, fmt.Errorf("list system catalog rows for parser: %w", err)
	}
	defer result.Close()

	rows := make([]model.SystemCatalogRow, 0)
	for result.Next() {
		var row model.SystemCatalogRow
		if err := result.Scan(&row.ID, &row.OrderID, &row.Position, &row.Code, &row.SystemName, &row.SystemURL, &row.SystemClass, &row.Curator, &row.ImportedAt); err != nil {
			return nil, fmt.Errorf("scan parser row: %w", err)
		}
		rows = append(rows, row)
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate parser rows: %w", err)
	}

	return rows, nil
}

func (r *SystemCatalogRepository) SaveParsed(ctx context.Context, systemName string, systemURL string, characteristics []model.SystemCharacteristic) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save parsed system: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	systemKey := systemMetadataKey(systemName)
	if systemKey == "" {
		return fmt.Errorf("parsed system name is empty")
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO nav_systems (system_key, system_name, system_url, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (system_key) DO UPDATE
		SET system_name = EXCLUDED.system_name, system_url = EXCLUDED.system_url, updated_at = NOW()
	`, systemKey, systemName, systemURL); err != nil {
		return fmt.Errorf("upsert parsed system metadata: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM nav_system_characteristics WHERE system_key = $1`, systemKey); err != nil {
		return fmt.Errorf("clear parsed characteristics: %w", err)
	}

	for _, characteristic := range characteristics {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO nav_system_characteristics (system_key, position, name, value)
			VALUES ($1, $2, $3, $4)
		`, systemKey, characteristic.Position, characteristic.Name, characteristic.Value); err != nil {
			return fmt.Errorf("insert parsed characteristic: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit parsed system: %w", err)
	}
	return nil
}

func (r *SystemCatalogRepository) ReplaceSystemTypes(ctx context.Context, systemTypes []model.SystemTypeOption) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace nav system types: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM nav_system_types`); err != nil {
		return fmt.Errorf("clear nav system types: %w", err)
	}
	for _, systemType := range systemTypes {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO nav_system_types (slug, name, image_url, image_content_type, image_data, position)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, systemType.Slug, systemType.Name, systemType.ImageURL, systemType.ImageContentType, systemType.ImageData, systemType.Position); err != nil {
			return fmt.Errorf("insert nav system type: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit nav system types: %w", err)
	}
	return nil
}

func (r *SystemCatalogRepository) SystemTypes(ctx context.Context) ([]model.SystemTypeOption, error) {
	result, err := r.db.QueryContext(ctx, `
		SELECT slug, name, image_url, position
		FROM nav_system_types
		ORDER BY position, name
	`)
	if err != nil {
		return nil, fmt.Errorf("list nav system types: %w", err)
	}
	defer result.Close()

	systemTypes := make([]model.SystemTypeOption, 0)
	for result.Next() {
		var systemType model.SystemTypeOption
		if err := result.Scan(&systemType.Slug, &systemType.Name, &systemType.ImageURL, &systemType.Position); err != nil {
			return nil, fmt.Errorf("scan nav system type: %w", err)
		}
		systemTypes = append(systemTypes, systemType)
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate nav system types: %w", err)
	}
	return systemTypes, nil
}

func (r *SystemCatalogRepository) SystemTypeImage(ctx context.Context, slug string) (model.SystemTypeImage, error) {
	var image model.SystemTypeImage
	if err := r.db.QueryRowContext(ctx, `
		SELECT image_content_type, image_data
		FROM nav_system_types
		WHERE slug = $1 AND image_data IS NOT NULL
	`, slug).Scan(&image.ContentType, &image.Data); err != nil {
		return model.SystemTypeImage{}, fmt.Errorf("load nav system type image: %w", err)
	}
	return image, nil
}

func (r *SystemCatalogRepository) NavParserSettings(ctx context.Context) (model.NavParserSettings, error) {
	var settings model.NavParserSettings
	if err := r.db.QueryRowContext(ctx, `
		SELECT update_interval_days, worker_count, request_timeout_seconds,
			retry_attempts, retry_delay_seconds, fallback_search, last_run_at
		FROM nav_parser_settings
		WHERE id = TRUE
	`).Scan(&settings.UpdateIntervalDays, &settings.WorkerCount, &settings.RequestTimeoutSecs,
		&settings.RetryAttempts, &settings.RetryDelaySecs, &settings.FallbackSearch, &settings.LastRunAt); err != nil {
		return model.NavParserSettings{}, fmt.Errorf("load nav parser settings: %w", err)
	}
	return settings, nil
}

func (r *SystemCatalogRepository) UpdateNavParserSettings(ctx context.Context, input model.NavParserSettings) (model.NavParserSettings, error) {
	var settings model.NavParserSettings
	if err := r.db.QueryRowContext(ctx, `
		UPDATE nav_parser_settings
		SET update_interval_days = $1, worker_count = $2, request_timeout_seconds = $3,
			retry_attempts = $4, retry_delay_seconds = $5, fallback_search = $6
		WHERE id = TRUE
		RETURNING update_interval_days, worker_count, request_timeout_seconds,
			retry_attempts, retry_delay_seconds, fallback_search, last_run_at
	`, input.UpdateIntervalDays, input.WorkerCount, input.RequestTimeoutSecs,
		input.RetryAttempts, input.RetryDelaySecs, input.FallbackSearch).Scan(
		&settings.UpdateIntervalDays, &settings.WorkerCount, &settings.RequestTimeoutSecs,
		&settings.RetryAttempts, &settings.RetryDelaySecs, &settings.FallbackSearch, &settings.LastRunAt); err != nil {
		return model.NavParserSettings{}, fmt.Errorf("update nav parser settings: %w", err)
	}
	return settings, nil
}

func (r *SystemCatalogRepository) MarkNavParserRun(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE nav_parser_settings
		SET last_run_at = NOW()
		WHERE id = TRUE
	`); err != nil {
		return fmt.Errorf("mark nav parser run: %w", err)
	}
	return nil
}

func (r *SystemCatalogRepository) SaveNavParserRun(ctx context.Context, run model.NavParserRun) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save nav parser run: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO nav_parser_runs (
			source, status, message, total, found, updated, failed, not_found, started_at, finished_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, run.Source, run.Status, run.Message, run.Total, run.Found, run.Updated, run.Failed, run.NotFound, run.StartedAt, run.FinishedAt).Scan(&run.ID); err != nil {
		return fmt.Errorf("insert nav parser run: %w", err)
	}

	for _, entry := range run.Logs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO nav_parser_run_logs (run_id, logged_at, level, message)
			VALUES ($1, $2, $3, $4)
		`, run.ID, entry.Time, entry.Level, entry.Message); err != nil {
			return fmt.Errorf("insert nav parser run log: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM nav_parser_runs
		WHERE id NOT IN (
			SELECT id
			FROM nav_parser_runs
			ORDER BY started_at DESC, id DESC
			LIMIT 5
		)
	`); err != nil {
		return fmt.Errorf("prune nav parser run history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit nav parser run: %w", err)
	}
	return nil
}

func (r *SystemCatalogRepository) NavParserRuns(ctx context.Context, limit int) ([]model.NavParserRun, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	result, err := r.db.QueryContext(ctx, `
		SELECT id, source, status, message, total, found, updated, failed, not_found, started_at, finished_at
		FROM nav_parser_runs
		ORDER BY started_at DESC, id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list nav parser runs: %w", err)
	}
	runs := make([]model.NavParserRun, 0)
	for result.Next() {
		var run model.NavParserRun
		if err := result.Scan(&run.ID, &run.Source, &run.Status, &run.Message, &run.Total, &run.Found, &run.Updated, &run.Failed, &run.NotFound, &run.StartedAt, &run.FinishedAt); err != nil {
			return nil, fmt.Errorf("scan nav parser run: %w", err)
		}
		run.Logs = make([]model.NavParserLogEntry, 0)
		runs = append(runs, run)
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate nav parser runs: %w", err)
	}
	if err := result.Close(); err != nil {
		return nil, fmt.Errorf("close nav parser runs: %w", err)
	}

	for index := range runs {
		logs, err := r.db.QueryContext(ctx, `
			SELECT logged_at, level, message
			FROM nav_parser_run_logs
			WHERE run_id = $1
			ORDER BY id
		`, runs[index].ID)
		if err != nil {
			return nil, fmt.Errorf("list nav parser run logs: %w", err)
		}
		for logs.Next() {
			var entry model.NavParserLogEntry
			if err := logs.Scan(&entry.Time, &entry.Level, &entry.Message); err != nil {
				_ = logs.Close()
				return nil, fmt.Errorf("scan nav parser run log: %w", err)
			}
			runs[index].Logs = append(runs[index].Logs, entry)
		}
		if err := logs.Err(); err != nil {
			_ = logs.Close()
			return nil, fmt.Errorf("iterate nav parser run logs: %w", err)
		}
		if err := logs.Close(); err != nil {
			return nil, fmt.Errorf("close nav parser run logs: %w", err)
		}
	}
	return runs, nil
}

func (r *SystemCatalogRepository) loadCharacteristics(ctx context.Context, rows []model.SystemCatalogRow) error {
	if len(rows) == 0 {
		return nil
	}

	byKey := make(map[string][]*model.SystemCatalogRow, len(rows))
	keys := make([]any, 0, len(rows))
	placeholders := make([]string, 0, len(rows))
	for index := range rows {
		key := systemMetadataKey(rows[index].SystemName)
		if _, exists := byKey[key]; !exists {
			keys = append(keys, key)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(keys)))
		}
		byKey[key] = append(byKey[key], &rows[index])
	}

	result, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT system_key, position, name, value
		FROM nav_system_characteristics
		WHERE system_key IN (%s)
		ORDER BY system_key, position
	`, strings.Join(placeholders, ",")), keys...)
	if err != nil {
		return fmt.Errorf("load system characteristics: %w", err)
	}
	defer result.Close()

	for result.Next() {
		var systemKey string
		var characteristic model.SystemCharacteristic
		if err := result.Scan(&systemKey, &characteristic.Position, &characteristic.Name, &characteristic.Value); err != nil {
			return fmt.Errorf("scan system characteristic: %w", err)
		}
		for _, row := range byKey[systemKey] {
			row.Characteristics = append(row.Characteristics, characteristic)
		}
	}
	if err := result.Err(); err != nil {
		return fmt.Errorf("iterate system characteristics: %w", err)
	}
	return nil
}

func systemMetadataKey(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}

func (r *SystemCatalogRepository) Stats(ctx context.Context, orderID int64) (model.SystemCatalogStats, error) {
	const query = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE system_class = 'Рекомендованная') AS recommended,
			COUNT(*) FILTER (WHERE system_class = 'Разрешенная') AS allowed,
			COUNT(*) FILTER (WHERE system_class = 'Запрещенная') AS forbidden,
			COUNT(DISTINCT NULLIF(curator, '')) AS curators
		FROM system_catalog
		WHERE ($1::BIGINT = 0 OR order_id = $1)
	`

	var stats model.SystemCatalogStats
	if err := r.db.QueryRowContext(ctx, query, orderID).Scan(&stats.Total, &stats.Recommended, &stats.Allowed, &stats.Forbidden, &stats.Curators); err != nil {
		return model.SystemCatalogStats{}, fmt.Errorf("load system catalog stats: %w", err)
	}

	return stats, nil
}

func (r *SystemCatalogRepository) Options(ctx context.Context, column string, orderID int64) ([]string, error) {
	if column != "system_class" && column != "curator" {
		return nil, fmt.Errorf("unsupported options column: %s", column)
	}

	result, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM system_catalog
		WHERE %s <> '' AND ($1::BIGINT = 0 OR order_id = $1)
		ORDER BY %s
	`, column, column, column), orderID)
	if err != nil {
		return nil, fmt.Errorf("load %s options: %w", column, err)
	}
	defer result.Close()

	options := []string{"Все"}
	for result.Next() {
		var option string
		if err := result.Scan(&option); err != nil {
			return nil, fmt.Errorf("scan %s option: %w", column, err)
		}
		options = append(options, option)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s options: %w", column, err)
	}

	return options, nil
}
