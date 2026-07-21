package repository

import (
	"context"
	"database/sql"
	"fmt"

	"tn/backend/internal/apperror"
	"tn/backend/internal/model"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) List(ctx context.Context) ([]model.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.ID, &order.Name, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) Create(ctx context.Context, name string) (model.Order, error) {
	var order model.Order
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO orders (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`, name).Scan(&order.ID, &order.Name, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return model.Order{}, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) Import(
	ctx context.Context,
	name string,
	classificationRows []model.ClassificationChange,
	systemCatalogRows []model.SystemCatalogRow,
) (model.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Order{}, fmt.Errorf("begin order workbook import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var order model.Order
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO orders (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`, name).Scan(&order.ID, &order.Name, &order.CreatedAt, &order.UpdatedAt); err != nil {
		return model.Order{}, fmt.Errorf("create imported order: %w", err)
	}

	classificationStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO classification_changes (
			order_id, position, system_name, system_url, construction_type, class_before, class_after
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return model.Order{}, fmt.Errorf("prepare imported classification rows: %w", err)
	}
	defer classificationStmt.Close()
	for _, row := range classificationRows {
		if _, err := classificationStmt.ExecContext(
			ctx, order.ID, row.Position, row.SystemName, row.SystemURL,
			row.ConstructionType, row.ClassBefore, row.ClassAfter,
		); err != nil {
			return model.Order{}, fmt.Errorf("insert imported classification row %q: %w", row.SystemName, err)
		}
	}

	catalogStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO system_catalog (
			order_id, position, code, system_name, system_url, system_class, curator, document_initialized
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING id
	`)
	if err != nil {
		return model.Order{}, fmt.Errorf("prepare imported system catalog rows: %w", err)
	}
	defer catalogStmt.Close()
	for _, row := range systemCatalogRows {
		var systemCatalogID int64
		if err := catalogStmt.QueryRowContext(
			ctx, order.ID, row.Position, row.Code, row.SystemName,
			row.SystemURL, row.SystemClass, row.Curator,
		).Scan(&systemCatalogID); err != nil {
			return model.Order{}, fmt.Errorf("insert imported system catalog row %q: %w", row.Code, err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO system_documents (order_id, system_catalog_id)
			VALUES ($1, $2)
		`, order.ID, systemCatalogID); err != nil {
			return model.Order{}, fmt.Errorf("create imported system document row %q: %w", row.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Order{}, fmt.Errorf("commit order workbook import: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) UpdateName(ctx context.Context, id int64, name string) (model.Order, error) {
	var order model.Order
	err := r.db.QueryRowContext(ctx, `
		UPDATE orders
		SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, created_at, updated_at
	`, id, name).Scan(&order.ID, &order.Name, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Order{}, apperror.New(apperror.NotFound, "order not found")
		}
		return model.Order{}, fmt.Errorf("update order name: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete order: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `LOCK TABLE orders IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return fmt.Errorf("lock orders before delete: %w", err)
	}

	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders`).Scan(&count); err != nil {
		return fmt.Errorf("count orders: %w", err)
	}
	if count <= 1 {
		return apperror.New(apperror.Conflict, "cannot delete the last order")
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM classification_changes WHERE order_id = $1`, id); err != nil {
		return fmt.Errorf("delete classification changes for order: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM system_catalog WHERE order_id = $1`, id); err != nil {
		return fmt.Errorf("delete system catalog for order: %w", err)
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted order: %w", err)
	}
	if affected == 0 {
		return apperror.New(apperror.NotFound, "order not found")
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit delete order: %w", err)
	}

	return nil
}
