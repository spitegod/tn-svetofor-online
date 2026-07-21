package service

import (
	"context"
	"strings"
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
	rows, err := service.parseSheet(context.Background(), workbook, "Таблица 2")
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

func TestSystemCatalogParseSheetRejectsUnknownClass(t *testing.T) {
	workbook := excelize.NewFile()
	defer workbook.Close()
	if err := workbook.SetCellValue("Sheet1", "A1", "Шифр"); err != nil {
		t.Fatal(err)
	}
	for cell, value := range map[string]string{
		"A2": "PK-1", "B2": "Тестовая система", "C2": "Неизвестная", "D2": "Куратор",
	} {
		if err := workbook.SetCellValue("Sheet1", cell, value); err != nil {
			t.Fatal(err)
		}
	}

	service := &SystemCatalogService{}
	_, err := service.parseSheet(context.Background(), workbook, "Sheet1")
	if err == nil || !strings.Contains(err.Error(), "недопустимый класс") {
		t.Fatalf("parseSheet() error = %v, want invalid class error", err)
	}
}
