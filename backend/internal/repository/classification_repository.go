package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tn/backend/internal/model"
)

type ClassificationRepository struct {
	db *sql.DB
}

func NewClassificationRepository(db *sql.DB) *ClassificationRepository {
	return &ClassificationRepository{db: db}
}

func (r *ClassificationRepository) ReplaceAll(ctx context.Context, orderID int64, rows []model.ClassificationChange) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace classification changes: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM classification_changes WHERE order_id = $1`, orderID); err != nil {
		return fmt.Errorf("clear classification changes: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO classification_changes (order_id, position, system_name, system_url, class_before, class_after)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("prepare classification insert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err = stmt.ExecContext(ctx, orderID, row.Position, row.SystemName, row.SystemURL, row.ClassBefore, row.ClassAfter); err != nil {
			return fmt.Errorf("insert classification change %q: %w", row.SystemName, err)
		}
	}

	if _, err = tx.ExecContext(ctx, `UPDATE orders SET updated_at = NOW() WHERE id = $1`, orderID); err != nil {
		return fmt.Errorf("touch order after classification import: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit classification changes: %w", err)
	}

	return nil
}

func (r *ClassificationRepository) List(ctx context.Context, filter model.ClassificationFilter) ([]model.ClassificationChange, error) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if filter.OrderID > 0 {
		args = append(args, filter.OrderID)
		clauses = append(clauses, fmt.Sprintf("order_id = $%d", len(args)))
	}

	if filter.Query != "" {
		args = append(args, "%"+strings.ToLower(filter.Query)+"%")
		clauses = append(clauses, fmt.Sprintf("LOWER(system_name) LIKE $%d", len(args)))
	}

	if filter.ClassBefore != "" {
		args = append(args, filter.ClassBefore)
		clauses = append(clauses, fmt.Sprintf("class_before = $%d", len(args)))
	}

	if filter.ClassAfter != "" {
		args = append(args, filter.ClassAfter)
		clauses = append(clauses, fmt.Sprintf("class_after = $%d", len(args)))
	}

	query := `
		SELECT id, order_id, position, system_name, system_url, class_before, class_after, imported_at
		FROM classification_changes
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY position, id"

	result, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list classification changes: %w", err)
	}
	defer result.Close()

	rows := make([]model.ClassificationChange, 0)
	for result.Next() {
		var row model.ClassificationChange
		if err := result.Scan(&row.ID, &row.OrderID, &row.Position, &row.SystemName, &row.SystemURL, &row.ClassBefore, &row.ClassAfter, &row.ImportedAt); err != nil {
			return nil, fmt.Errorf("scan classification change: %w", err)
		}
		rows = append(rows, row)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate classification changes: %w", err)
	}

	return rows, nil
}

func (r *ClassificationRepository) Stats(ctx context.Context, orderID int64) (model.ClassificationStats, error) {
	const query = `
		SELECT
			COUNT(*) FILTER (WHERE class_before = 'Новая система') AS added_systems,
			COUNT(*) FILTER (WHERE class_after = 'Рекомендованная') AS recommended,
			COUNT(*) FILTER (WHERE class_after = 'Разрешенная') AS allowed,
			COUNT(*) FILTER (WHERE class_before <> 'Новая система') AS classification_changes
		FROM classification_changes
		WHERE ($1::BIGINT = 0 OR order_id = $1)
	`

	var stats model.ClassificationStats
	if err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&stats.AddedSystems,
		&stats.Recommended,
		&stats.Allowed,
		&stats.ClassificationChanges,
	); err != nil {
		return model.ClassificationStats{}, fmt.Errorf("load classification stats: %w", err)
	}

	return stats, nil
}

func (r *ClassificationRepository) Options(ctx context.Context, column string, orderID int64) ([]string, error) {
	if column != "class_before" && column != "class_after" {
		return nil, fmt.Errorf("unsupported options column: %s", column)
	}

	result, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM classification_changes
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
