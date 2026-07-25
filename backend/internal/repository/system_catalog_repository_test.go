package repository

import (
	"context"
	"errors"
	"testing"

	"tn/backend/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSystemCatalogReplaceAllRollsBackWhenPreservationFails(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WithArgs(int64(6_200_000_001)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("CREATE TEMPORARY TABLE preserved_system_document_state").WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()

	err = NewSystemCatalogRepository(database).ReplaceAll(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected replacement to fail")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSystemCatalogReplaceAllCastsDocumentForeignKeys(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	const orderID int64 = 7
	const catalogID int64 = 42

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").
		WithArgs(int64(6_200_000_000) + orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("CREATE TEMPORARY TABLE preserved_system_document_state").
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM system_catalog").
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectPrepare("INSERT INTO system_catalog").
		ExpectQuery().
		WithArgs(orderID, 1, "ПК-1", "Система", "", "Разрешенная", "Куратор").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(catalogID))
	mock.ExpectExec(`(?s)INSERT INTO system_documents.*SELECT \$1::BIGINT, \$2::BIGINT.*UNION ALL.*SELECT \$1::BIGINT, \$2::BIGINT`).
		WithArgs(orderID, catalogID, "ПК-1", "Система").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE orders SET updated_at").
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rows := []model.SystemCatalogRow{{
		Position:    1,
		Code:        "ПК-1",
		SystemName:  "Система",
		SystemClass: "Разрешенная",
		Curator:     "Куратор",
	}}
	if err := NewSystemCatalogRepository(database).ReplaceAll(context.Background(), orderID, rows); err != nil {
		t.Fatalf("replace catalog: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
