package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"tn/backend/internal/model"
	"tn/backend/internal/repository"

	"github.com/xuri/excelize/v2"
)

type SystemCatalogService struct {
	repo *repository.SystemCatalogRepository
}

func NewSystemCatalogService(repo *repository.SystemCatalogRepository) *SystemCatalogService {
	return &SystemCatalogService{repo: repo}
}

func (s *SystemCatalogService) List(ctx context.Context, filter model.SystemCatalogFilter) (model.SystemCatalogList, error) {
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return model.SystemCatalogList{}, err
	}

	stats, err := s.repo.Stats(ctx, filter.OrderID)
	if err != nil {
		return model.SystemCatalogList{}, err
	}

	classOptions, err := s.repo.Options(ctx, "system_class", filter.OrderID)
	if err != nil {
		return model.SystemCatalogList{}, err
	}

	curatorOptions, err := s.repo.Options(ctx, "curator", filter.OrderID)
	if err != nil {
		return model.SystemCatalogList{}, err
	}
	systemTypes, err := s.repo.SystemTypes(ctx)
	if err != nil {
		return model.SystemCatalogList{}, err
	}

	return model.SystemCatalogList{
		Rows:           rows,
		Stats:          stats,
		ClassOptions:   classOptions,
		CuratorOptions: curatorOptions,
		SystemTypes:    systemTypes,
	}, nil
}

func (s *SystemCatalogService) Import(ctx context.Context, orderID int64, file io.Reader) (model.SystemCatalogList, error) {
	if orderID <= 0 {
		orderID = 1
	}

	spreadsheet, err := excelize.OpenReader(file)
	if err != nil {
		return model.SystemCatalogList{}, fmt.Errorf("open excel file: %w", err)
	}
	defer spreadsheet.Close()

	sheetName := spreadsheet.GetSheetName(0)
	if sheetName == "" {
		return model.SystemCatalogList{}, fmt.Errorf("excel file has no sheets")
	}

	excelRows, err := spreadsheet.GetRows(sheetName)
	if err != nil {
		return model.SystemCatalogList{}, fmt.Errorf("read excel rows: %w", err)
	}

	rows := make([]model.SystemCatalogRow, 0, len(excelRows))
	for index, excelRow := range excelRows {
		if index == 0 || len(excelRow) < 4 {
			continue
		}

		code := normalizeCell(excelRow[0])
		systemName := normalizeCell(excelRow[1])
		systemClass := normalizeStatus(excelRow[2])
		curator := normalizeCell(excelRow[3])
		if code == "" || systemName == "" || systemClass == "" {
			continue
		}

		rows = append(rows, model.SystemCatalogRow{
			Position:    len(rows) + 1,
			Code:        code,
			SystemName:  systemName,
			SystemClass: systemClass,
			Curator:     curator,
		})
	}

	if len(rows) == 0 {
		return model.SystemCatalogList{}, fmt.Errorf("excel file has no system catalog rows")
	}

	if err := s.repo.ReplaceAll(ctx, orderID, rows); err != nil {
		return model.SystemCatalogList{}, err
	}

	return s.List(ctx, model.SystemCatalogFilter{OrderID: orderID})
}

func (s *SystemCatalogService) Export(ctx context.Context, filter model.SystemCatalogFilter) ([]byte, error) {
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer file.Close()

	sheet := "Таблица 2"
	file.SetSheetName("Sheet1", sheet)
	_ = file.SetCellValue(sheet, "A1", "Шифр")
	_ = file.SetCellValue(sheet, "B1", "Название")
	_ = file.SetCellValue(sheet, "C1", "Класс")
	_ = file.SetCellValue(sheet, "D1", "Куратор")

	for index, row := range rows {
		excelRow := index + 2
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), row.Code)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), row.SystemName)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", excelRow), row.SystemClass)
		_ = file.SetCellValue(sheet, fmt.Sprintf("D%d", excelRow), row.Curator)
	}

	_ = file.SetColWidth(sheet, "A", "A", 18)
	_ = file.SetColWidth(sheet, "B", "B", 48)
	_ = file.SetColWidth(sheet, "C", "D", 22)

	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, fmt.Errorf("write excel export: %w", err)
	}

	return buffer.Bytes(), nil
}
