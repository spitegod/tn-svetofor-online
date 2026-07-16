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
		INSERT INTO classification_changes (order_id, position, system_name, system_url, construction_type, class_before, class_after)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("prepare classification insert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err = stmt.ExecContext(ctx, orderID, row.Position, row.SystemName, row.SystemURL, row.ConstructionType, row.ClassBefore, row.ClassAfter); err != nil {
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

func (r *ClassificationRepository) Update(ctx context.Context, id int64, orderID int64, row model.ClassificationChange) (model.ClassificationChange, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ClassificationChange{}, fmt.Errorf("begin classification change update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var updated model.ClassificationChange
	err = tx.QueryRowContext(ctx, `
		UPDATE classification_changes
		SET system_name = $3, class_before = $4, class_after = $5
		WHERE id = $1 AND order_id = $2
		RETURNING id, order_id, position, system_name, system_url, class_before, class_after, imported_at
	`, id, orderID, row.SystemName, row.ClassBefore, row.ClassAfter).Scan(
		&updated.ID, &updated.OrderID, &updated.Position, &updated.SystemName, &updated.SystemURL,
		&updated.ClassBefore, &updated.ClassAfter, &updated.ImportedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.ClassificationChange{}, fmt.Errorf("classification change not found")
		}
		return model.ClassificationChange{}, fmt.Errorf("update classification change: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `UPDATE orders SET updated_at = NOW() WHERE id = $1`, orderID); err != nil {
		return model.ClassificationChange{}, fmt.Errorf("touch order after classification update: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return model.ClassificationChange{}, fmt.Errorf("commit classification change update: %w", err)
	}
	return updated, nil
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

	if filter.ConstructionType != "" {
		args = append(args, filter.ConstructionType)
		clauses = append(clauses, fmt.Sprintf("construction_type = $%d", len(args)))
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
		SELECT id, order_id, position, system_name, system_url, construction_type, class_before, class_after, imported_at
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
		if err := result.Scan(&row.ID, &row.OrderID, &row.Position, &row.SystemName, &row.SystemURL, &row.ConstructionType, &row.ClassBefore, &row.ClassAfter, &row.ImportedAt); err != nil {
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
		WITH current_order AS (
			SELECT id, created_at
			FROM orders
			WHERE id = $1
		),
		previous_order AS (
			SELECT candidate.id
			FROM orders candidate
			CROSS JOIN current_order current
			WHERE candidate.created_at < current.created_at
				OR (candidate.created_at = current.created_at AND candidate.id < current.id)
			ORDER BY candidate.created_at DESC, candidate.id DESC
			LIMIT 1
		),
		current_systems AS (
			SELECT DISTINCT ON (LOWER(TRIM(system_name)))
				LOWER(TRIM(system_name)) AS system_key,
				class_before,
				class_after
			FROM classification_changes
			WHERE order_id = $1
			ORDER BY LOWER(TRIM(system_name)), position DESC, id DESC
		),
		previous_systems AS (
			SELECT DISTINCT LOWER(TRIM(change.system_name)) AS system_key
			FROM classification_changes change
			WHERE change.order_id = (SELECT id FROM previous_order)
		),
		compared AS (
			SELECT current.*, previous.system_key IS NULL AS is_new
			FROM current_systems current
			LEFT JOIN previous_systems previous USING (system_key)
		)
		SELECT
			COUNT(*) FILTER (WHERE is_new) AS added_systems,
			COUNT(*) FILTER (WHERE class_after = 'Рекомендованная') AS recommended,
			COUNT(*) FILTER (WHERE class_after = 'Разрешенная') AS allowed,
			COUNT(*) FILTER (WHERE class_before <> 'Новая система') AS classification_changes
		FROM compared
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
