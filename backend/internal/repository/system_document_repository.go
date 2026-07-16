package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tn/backend/internal/model"
)

type SystemDocumentRepository struct {
	db *sql.DB
}

func NewSystemDocumentRepository(db *sql.DB) *SystemDocumentRepository {
	return &SystemDocumentRepository{db: db}
}

func (r *SystemDocumentRepository) List(ctx context.Context, filter model.SystemDocumentFilter) ([]model.SystemDocumentRow, error) {
	clauses := []string{"d.order_id = $1"}
	args := []any{filter.OrderID}
	if filter.Query != "" {
		args = append(args, "%"+strings.ToLower(filter.Query)+"%")
		clauses = append(clauses, fmt.Sprintf("LOWER(s.system_name) LIKE $%d", len(args)))
	}
	if filter.SystemClass != "" {
		args = append(args, filter.SystemClass)
		clauses = append(clauses, fmt.Sprintf("s.system_class = $%d", len(args)))
	}
	if filter.Curator != "" {
		args = append(args, filter.Curator)
		clauses = append(clauses, fmt.Sprintf("s.curator = $%d", len(args)))
	}
	if filter.ComparisonOnly {
		clauses = append(clauses, "d.comparison_selected = TRUE")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.order_id, o.name, d.system_catalog_id, s.position, s.code,
			s.system_name, COALESCE(NULLIF(nav.system_url, ''), s.system_url), s.system_class, s.curator, d.comparison_selected, d.comment,
			d.created_at, d.updated_at
		FROM system_documents d
		JOIN system_catalog s ON s.id = d.system_catalog_id
		LEFT JOIN nav_systems nav
			ON nav.system_key = LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g'))
		JOIN orders o ON o.id = d.order_id
		WHERE `+strings.Join(clauses, " AND ")+`
		ORDER BY CASE WHEN BTRIM(d.comment) <> '' THEN 0 ELSE 1 END, LOWER(s.system_name), s.position, d.id
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list system documents: %w", err)
	}
	defer rows.Close()

	result := make([]model.SystemDocumentRow, 0)
	for rows.Next() {
		var row model.SystemDocumentRow
		if err := rows.Scan(&row.ID, &row.OrderID, &row.OrderName, &row.SystemCatalogID, &row.Position, &row.Code,
			&row.SystemName, &row.SystemURL, &row.SystemClass, &row.Curator, &row.ComparisonSelected, &row.Comment,
			&row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan system document: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate system documents: %w", err)
	}
	if err := r.loadCharacteristics(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SystemDocumentRepository) History(ctx context.Context, code string, systemName string) ([]model.SystemDocumentRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.order_id, o.name, d.system_catalog_id, s.position, s.code,
			s.system_name, COALESCE(NULLIF(nav.system_url, ''), s.system_url), s.system_class, s.curator, d.comparison_selected, d.comment,
			d.created_at, d.updated_at
		FROM system_documents d
		JOIN system_catalog s ON s.id = d.system_catalog_id
		LEFT JOIN nav_systems nav
			ON nav.system_key = LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g'))
		JOIN orders o ON o.id = d.order_id
		WHERE s.code = $1 AND s.system_name = $2
		ORDER BY o.created_at DESC, o.id DESC
	`, code, systemName)
	if err != nil {
		return nil, fmt.Errorf("load system document history: %w", err)
	}
	defer rows.Close()

	result := make([]model.SystemDocumentRow, 0)
	for rows.Next() {
		var row model.SystemDocumentRow
		if err := rows.Scan(&row.ID, &row.OrderID, &row.OrderName, &row.SystemCatalogID, &row.Position, &row.Code,
			&row.SystemName, &row.SystemURL, &row.SystemClass, &row.Curator, &row.ComparisonSelected, &row.Comment,
			&row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan system document history: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate system document history: %w", err)
	}
	if err := r.loadCharacteristics(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SystemDocumentRepository) loadCharacteristics(ctx context.Context, rows []model.SystemDocumentRow) error {
	if len(rows) == 0 {
		return nil
	}
	bySystemKey := make(map[string][]*model.SystemDocumentRow, len(rows))
	args := make([]any, 0, len(rows))
	placeholders := make([]string, 0, len(rows))
	for index := range rows {
		key := systemMetadataKey(rows[index].SystemName)
		if _, exists := bySystemKey[key]; !exists {
			args = append(args, key)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		bySystemKey[key] = append(bySystemKey[key], &rows[index])
	}
	result, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT system_key, position, name, value
		FROM nav_system_characteristics
		WHERE system_key IN (%s)
		ORDER BY system_key, position
	`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return fmt.Errorf("load system document characteristics: %w", err)
	}
	defer result.Close()
	for result.Next() {
		var systemKey string
		var characteristic model.SystemCharacteristic
		if err := result.Scan(&systemKey, &characteristic.Position, &characteristic.Name, &characteristic.Value); err != nil {
			return fmt.Errorf("scan system document characteristic: %w", err)
		}
		for _, row := range bySystemKey[systemKey] {
			row.Characteristics = append(row.Characteristics, characteristic)
		}
	}
	if err := result.Err(); err != nil {
		return fmt.Errorf("iterate system document characteristics: %w", err)
	}
	return nil
}

func (r *SystemDocumentRepository) UpdateComment(ctx context.Context, id int64, orderID int64, comment string) (model.SystemDocumentRow, error) {
	var row model.SystemDocumentRow
	err := r.db.QueryRowContext(ctx, `
		WITH updated AS (
			UPDATE system_documents
			SET comment = $3, updated_at = NOW()
			WHERE id = $1 AND order_id = $2
			RETURNING *
		)
		SELECT d.id, d.order_id, o.name, d.system_catalog_id, s.position, s.code,
			s.system_name, COALESCE(NULLIF(nav.system_url, ''), s.system_url), s.system_class, s.curator, d.comparison_selected, d.comment,
			d.created_at, d.updated_at
		FROM updated d
		JOIN system_catalog s ON s.id = d.system_catalog_id
		LEFT JOIN nav_systems nav
			ON nav.system_key = LOWER(REGEXP_REPLACE(BTRIM(s.system_name), '\s+', ' ', 'g'))
		JOIN orders o ON o.id = d.order_id
	`, id, orderID, comment).Scan(&row.ID, &row.OrderID, &row.OrderName, &row.SystemCatalogID, &row.Position, &row.Code,
		&row.SystemName, &row.SystemURL, &row.SystemClass, &row.Curator, &row.ComparisonSelected, &row.Comment,
		&row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.SystemDocumentRow{}, fmt.Errorf("system document not found")
		}
		return model.SystemDocumentRow{}, fmt.Errorf("update system document comment: %w", err)
	}
	return row, nil
}

func (r *SystemDocumentRepository) UpdateComparison(ctx context.Context, id int64, orderID int64, selected bool) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE system_documents
		SET comparison_selected = $3, updated_at = NOW()
		WHERE id = $1 AND order_id = $2
	`, id, orderID, selected)
	if err != nil {
		return fmt.Errorf("update system document comparison selection: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check system document comparison selection: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("system document not found")
	}
	return nil
}

func (r *SystemDocumentRepository) UpdateComparisonBulk(ctx context.Context, orderID int64, allOrders bool, selected bool, systems []model.SystemDocumentKey) error {
	if len(systems) == 0 {
		return nil
	}
	args := []any{selected}
	clauses := make([]string, 0, len(systems))
	for _, system := range systems {
		args = append(args, system.Code, system.SystemName)
		clauses = append(clauses, fmt.Sprintf("(s.code = $%d AND s.system_name = $%d)", len(args)-1, len(args)))
	}
	orderClause := ""
	if !allOrders {
		args = append(args, orderID)
		orderClause = fmt.Sprintf(" AND d.order_id = $%d", len(args))
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE system_documents d
		SET comparison_selected = $1, updated_at = NOW()
		FROM system_catalog s
		WHERE s.id = d.system_catalog_id
			AND (`+strings.Join(clauses, " OR ")+`)`+orderClause,
		args...,
	)
	if err != nil {
		return fmt.Errorf("bulk update comparison selection: %w", err)
	}
	return nil
}

func (r *SystemDocumentRepository) Delete(ctx context.Context, id int64, orderID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete system document: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `DELETE FROM system_documents WHERE id = $1 AND order_id = $2`, id, orderID)
	if err != nil {
		return fmt.Errorf("delete system document: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted system document: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("system document not found")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE orders SET updated_at = NOW() WHERE id = $1`, orderID); err != nil {
		return fmt.Errorf("touch order after system document delete: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit system document delete: %w", err)
	}
	return nil
}

func (r *SystemDocumentRepository) Stats(ctx context.Context, orderID int64) (model.SystemCatalogStats, error) {
	var stats model.SystemCatalogStats
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE s.system_class = 'Рекомендованная'),
			COUNT(*) FILTER (WHERE s.system_class = 'Разрешенная'),
			COUNT(*) FILTER (WHERE s.system_class = 'Запрещенная'),
			COUNT(DISTINCT NULLIF(s.curator, ''))
		FROM system_documents d
		JOIN system_catalog s ON s.id = d.system_catalog_id
		WHERE d.order_id = $1
	`, orderID).Scan(&stats.Total, &stats.Recommended, &stats.Allowed, &stats.Forbidden, &stats.Curators)
	if err != nil {
		return model.SystemCatalogStats{}, fmt.Errorf("load system document stats: %w", err)
	}
	return stats, nil
}

func (r *SystemDocumentRepository) Options(ctx context.Context, column string, orderID int64) ([]string, error) {
	if column != "system_class" && column != "curator" {
		return nil, fmt.Errorf("unsupported system document option: %s", column)
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT DISTINCT s.%s
		FROM system_documents d
		JOIN system_catalog s ON s.id = d.system_catalog_id
		WHERE d.order_id = $1 AND s.%s <> ''
		ORDER BY s.%s
	`, column, column, column), orderID)
	if err != nil {
		return nil, fmt.Errorf("load system document %s options: %w", column, err)
	}
	defer rows.Close()
	options := []string{"Все"}
	for rows.Next() {
		var option string
		if err := rows.Scan(&option); err != nil {
			return nil, fmt.Errorf("scan system document option: %w", err)
		}
		options = append(options, option)
	}
	return options, rows.Err()
}
