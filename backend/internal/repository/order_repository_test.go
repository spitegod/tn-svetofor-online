package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"tn/backend/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestOrderDeleteRollsBackWhenLastOrderCannotBeDeleted(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	mock.ExpectBegin()
	mock.ExpectExec("LOCK TABLE orders").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM orders").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	err = NewOrderRepository(database).Delete(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "cannot delete the last order") {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrderImportRollsBackAllTablesOnFailure(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO orders").WithArgs("Order").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).AddRow(2, "Order", now, now),
	)
	classification := mock.ExpectPrepare("INSERT INTO classification_changes")
	classification.ExpectExec().WithArgs(
		int64(2), 1, "System", "", "Type", "Новая система", "Разрешенная",
	).WillReturnResult(sqlmock.NewResult(1, 1))
	catalog := mock.ExpectPrepare("INSERT INTO system_catalog")
	catalog.ExpectQuery().WithArgs(
		int64(2), 1, "PK-1", "System", "", "Разрешенная", "Curator",
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec("INSERT INTO system_documents").WithArgs(int64(2), int64(10)).WillReturnError(errors.New("write failed"))
	mock.ExpectRollback()

	_, err = NewOrderRepository(database).Import(
		context.Background(),
		"Order",
		[]model.ClassificationChange{{Position: 1, SystemName: "System", ConstructionType: "Type", ClassBefore: "Новая система", ClassAfter: "Разрешенная"}},
		[]model.SystemCatalogRow{{Position: 1, Code: "PK-1", SystemName: "System", SystemClass: "Разрешенная", Curator: "Curator"}},
	)
	if err == nil {
		t.Fatal("expected import to fail")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
