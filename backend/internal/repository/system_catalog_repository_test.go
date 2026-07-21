package repository

import (
	"context"
	"errors"
	"testing"

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
