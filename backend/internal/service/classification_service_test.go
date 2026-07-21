package service

import (
	"context"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestNormalizeConstructionType(t *testing.T) {
	tests := map[string]string{
		"ИЖС": "Индивидуальное жилищное строительство",
		"транспортное строительство": "Транспортное и дорожное строительство",
		"Специальные сооружения":     "Специальные сооружения",
		"Спец. сооружения":           "Специальные сооружения",
		"СС":                         "Специальные сооружения",
		"неизвестный сегмент":        unassignedConstructionType,
	}
	for value, want := range tests {
		if got := normalizeConstructionType(value); got != want {
			t.Errorf("normalizeConstructionType(%q) = %q, want %q", value, got, want)
		}
	}
}

func TestKnownNAVSystemDataNormalizesSystemPrefix(t *testing.T) {
	got, found := knownNAVSystemData("СИСТЕМА ТН-КРОВЛЯ СОЛИД КЕРАМЗИТ")
	if !found {
		t.Fatal("knownNAVSystemData() did not find system")
	}
	if got.ConstructionType != "Промышленное и гражданское строительство" {
		t.Fatalf("knownNAVSystemData() construction type = %q", got.ConstructionType)
	}
}

func TestClassificationImportParsingDoesNotRequireNAVNetworkLookup(t *testing.T) {
	workbook := excelize.NewFile()
	defer workbook.Close()
	for cell, value := range map[string]string{
		"A1": "Название системы", "B1": "Класс", "B2": "было", "C2": "стало",
		"A3": "Неизвестная система", "B3": "Новая система", "C3": "Разрешенная",
	} {
		if err := workbook.SetCellValue("Sheet1", cell, value); err != nil {
			t.Fatal(err)
		}
	}

	rows, err := NewClassificationService(nil).parseSheet(context.Background(), workbook, "Sheet1")
	if err != nil {
		t.Fatalf("parse classification: %v", err)
	}
	if len(rows) != 1 || rows[0].ConstructionType != unassignedConstructionType {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}
