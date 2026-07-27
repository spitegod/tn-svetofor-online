package repository

import (
	"context"
	"testing"
	"time"

	"tn/backend/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestClassificationUpdateReturnsConstructionType(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE classification_changes").WithArgs(
		int64(10), int64(2), "System", "Новая система", "Разрешенная",
	).WillReturnRows(sqlmock.NewRows([]string{
		"id", "order_id", "position", "system_name", "system_url", "construction_type", "class_before", "class_after", "imported_at",
	}).AddRow(10, 2, 1, "System", "", "Специальные сооружения", "Новая система", "Разрешенная", now))
	mock.ExpectExec("UPDATE orders").WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	updated, err := NewClassificationRepository(database).Update(context.Background(), 10, 2, model.ClassificationChange{
		SystemName: "System", ClassBefore: "Новая система", ClassAfter: "Разрешенная",
	})
	if err != nil {
		t.Fatalf("update classification: %v", err)
	}
	if updated.ConstructionType != "Специальные сооружения" {
		t.Fatalf("unexpected construction type %q", updated.ConstructionType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClassificationStatsCountsNewSystemMarker(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	mock.ExpectQuery(`COUNT\(\*\) FILTER \(WHERE class_before = 'Новая система'\) AS added_systems`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"added_systems", "recommended", "allowed", "classification_changes",
		}).AddRow(12, 7, 5, 3))

	stats, err := NewClassificationRepository(database).Stats(context.Background(), 5)
	if err != nil {
		t.Fatalf("load classification stats: %v", err)
	}
	if stats.AddedSystems != 12 {
		t.Fatalf("unexpected added systems count %d", stats.AddedSystems)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
