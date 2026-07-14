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
	}
	documentStates := make(map[string]documentState)
	existing, queryErr := tx.QueryContext(ctx, `
		SELECT s.code, s.system_name, d.comment, d.comparison_selected
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
		if err = existing.Scan(&code, &name, &state.comment, &state.comparisonSelected); err != nil {
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
			INSERT INTO system_documents (order_id, system_catalog_id, comment, comparison_selected)
			VALUES ($1, $2, $3, $4)
		`, orderID, systemCatalogID, state.comment, state.comparisonSelected); err != nil {
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

func (r *SystemCatalogRepository) List(ctx context.Context, filter model.SystemCatalogFilter) ([]model.SystemCatalogRow, error) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if filter.OrderID > 0 {
		args = append(args, filter.OrderID)
		clauses = append(clauses, fmt.Sprintf("order_id = $%d", len(args)))
	}

	if filter.Query != "" {
		args = append(args, "%"+strings.ToLower(filter.Query)+"%")
		clauses = append(clauses, fmt.Sprintf("(LOWER(system_name) LIKE $%d OR LOWER(code) LIKE $%d)", len(args), len(args)))
	}

	if filter.SystemClass != "" {
		args = append(args, filter.SystemClass)
		clauses = append(clauses, fmt.Sprintf("system_class = $%d", len(args)))
	}

	if filter.Curator != "" {
		args = append(args, filter.Curator)
		clauses = append(clauses, fmt.Sprintf("curator = $%d", len(args)))
	}

	query := `
		SELECT id, order_id, position, code, system_name, system_url, system_class, curator, imported_at
		FROM system_catalog
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY position, id"

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

func (r *SystemCatalogRepository) ParserRows(ctx context.Context, orderID int64) ([]model.SystemCatalogRow, error) {
	result, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, position, code, system_name, system_url, system_class, curator, imported_at
		FROM system_catalog
		WHERE order_id = $1
		ORDER BY position, id
	`, orderID)
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

func (r *SystemCatalogRepository) SaveParsed(ctx context.Context, systemID int64, systemURL string, characteristics []model.SystemCharacteristic) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save parsed system: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `UPDATE system_catalog SET system_url = $2 WHERE id = $1`, systemID, systemURL); err != nil {
		return fmt.Errorf("update parsed system url: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM system_characteristics WHERE system_catalog_id = $1`, systemID); err != nil {
		return fmt.Errorf("clear parsed characteristics: %w", err)
	}

	for _, characteristic := range characteristics {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO system_characteristics (system_catalog_id, position, name, value)
			VALUES ($1, $2, $3, $4)
		`, systemID, characteristic.Position, characteristic.Name, characteristic.Value); err != nil {
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
			INSERT INTO nav_system_types (slug, name, position)
			VALUES ($1, $2, $3)
		`, systemType.Slug, systemType.Name, systemType.Position); err != nil {
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
		SELECT slug, name, position
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
		if err := result.Scan(&systemType.Slug, &systemType.Name, &systemType.Position); err != nil {
			return nil, fmt.Errorf("scan nav system type: %w", err)
		}
		systemTypes = append(systemTypes, systemType)
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate nav system types: %w", err)
	}
	return systemTypes, nil
}

func (r *SystemCatalogRepository) loadCharacteristics(ctx context.Context, rows []model.SystemCatalogRow) error {
	if len(rows) == 0 {
		return nil
	}

	byID := make(map[int64]*model.SystemCatalogRow, len(rows))
	ids := make([]any, 0, len(rows))
	placeholders := make([]string, 0, len(rows))
	for index := range rows {
		byID[rows[index].ID] = &rows[index]
		ids = append(ids, rows[index].ID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}

	result, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT system_catalog_id, position, name, value
		FROM system_characteristics
		WHERE system_catalog_id IN (%s)
		ORDER BY system_catalog_id, position, id
	`, strings.Join(placeholders, ",")), ids...)
	if err != nil {
		return fmt.Errorf("load system characteristics: %w", err)
	}
	defer result.Close()

	for result.Next() {
		var systemID int64
		var characteristic model.SystemCharacteristic
		if err := result.Scan(&systemID, &characteristic.Position, &characteristic.Name, &characteristic.Value); err != nil {
			return fmt.Errorf("scan system characteristic: %w", err)
		}
		if row := byID[systemID]; row != nil {
			row.Characteristics = append(row.Characteristics, characteristic)
		}
	}
	if err := result.Err(); err != nil {
		return fmt.Errorf("iterate system characteristics: %w", err)
	}
	return nil
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
