package service

import (
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestSystemCatalogParseSheetKeepsRowsWithoutCode(t *testing.T) {
	workbook := excelize.NewFile()
	defer workbook.Close()
	workbook.SetSheetName("Sheet1", "Таблица 2")
	for cell, value := range map[string]string{
		"A1": "Шифр", "B1": "Название", "C1": "Класс", "D1": "Куратор",
		"B2": "ТН-ГЕО Полигон Фрост", "C2": "Рекомендованная", "D2": "Хомяков Я.",
	} {
		if err := workbook.SetCellValue("Таблица 2", cell, value); err != nil {
			t.Fatal(err)
		}
	}

	service := &SystemCatalogService{}
	rows, err := service.parseSheet(workbook, "Таблица 2")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("parseSheet() rows = %d, want 1", len(rows))
	}
	if rows[0].Code != "" {
		t.Fatalf("parseSheet() code = %q, want empty", rows[0].Code)
	}
	if rows[0].SystemName != "ТН-ГЕО Полигон Фрост" {
		t.Fatalf("parseSheet() system name = %q", rows[0].SystemName)
	}
}
