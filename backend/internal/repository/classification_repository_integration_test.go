package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"tn/backend/internal/db"
	"tn/backend/internal/model"
)

func TestClassificationListUsesSharedParserMetadata(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	defer database.Close()
	if err := db.Migrate(ctx, database); err != nil {
		t.Fatalf("migrate integration database: %v", err)
	}

	if _, err := database.ExecContext(ctx, `
		INSERT INTO classification_changes (
			order_id, position, system_name, construction_type, class_before, class_after
		) VALUES (1, 1, 'ТН-ТЕСТ', 'Тип не присвоен', 'Новая система', 'Разрешенная');
		INSERT INTO nav_systems (system_key, system_name, system_url)
		VALUES ('тн-тест', 'ТН-ТЕСТ', 'https://nav.tn.ru/systems/test/tn-test/');
		INSERT INTO nav_system_characteristics (system_key, position, name, value)
		VALUES ('тн-тест', 1, 'Сегмент строительства', 'Специальные сооружения');
	`); err != nil {
		t.Fatalf("seed integration database: %v", err)
	}

	rows, err := NewClassificationRepository(database).List(ctx, model.ClassificationFilter{
		OrderID: 1, Query: "ТН-ТЕСТ", ConstructionType: "Специальные сооружения",
	})
	if err != nil {
		t.Fatalf("list classification: %v", err)
	}
	if len(rows) != 1 || rows[0].ConstructionType != "Специальные сооружения" || rows[0].SystemURL == "" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}
