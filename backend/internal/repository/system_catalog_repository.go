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

	if _, err = tx.ExecContext(ctx, `DELETE FROM system_catalog WHERE order_id = $1`, orderID); err != nil {
		return fmt.Errorf("clear system catalog: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO system_catalog (order_id, position, code, system_name, system_url, system_class, curator)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("prepare system catalog insert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err = stmt.ExecContext(ctx, orderID, row.Position, row.Code, row.SystemName, row.SystemURL, row.SystemClass, row.Curator); err != nil {
			return fmt.Errorf("insert system catalog row %q: %w", row.Code, err)
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

	return rows, nil
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
