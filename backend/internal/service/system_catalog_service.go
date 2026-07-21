package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"tn/backend/internal/apperror"
	"tn/backend/internal/model"

	"github.com/xuri/excelize/v2"
)

type SystemCatalogService struct {
	repo systemCatalogRepository
}

type systemCatalogRepository interface {
	List(context.Context, model.SystemCatalogFilter) ([]model.SystemCatalogRow, error)
	Stats(context.Context, int64) (model.SystemCatalogStats, error)
	Options(context.Context, string, int64) ([]string, error)
	SystemTypes(context.Context) ([]model.SystemTypeOption, error)
	ReplaceAll(context.Context, int64, []model.SystemCatalogRow) error
	Update(context.Context, int64, int64, model.SystemCatalogRow) (model.SystemCatalogRow, error)
}

func NewSystemCatalogService(repo systemCatalogRepository) *SystemCatalogService {
	return &SystemCatalogService{repo: repo}
}

func (s *SystemCatalogService) List(ctx context.Context, filter model.SystemCatalogFilter) (model.SystemCatalogList, error) {
	if err := validateSystemCatalogFilter(filter); err != nil {
		return model.SystemCatalogList{}, err
	}
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

	spreadsheet, err := excelize.OpenReader(file, spreadsheetOpenOptions())
	if err != nil {
		return model.SystemCatalogList{}, apperror.Wrap(apperror.Validation, fmt.Errorf("open excel file: %w", err))
	}
	defer spreadsheet.Close()

	sheetName := spreadsheet.GetSheetName(0)
	if sheetName == "" {
		return model.SystemCatalogList{}, apperror.New(apperror.Validation, "excel file has no sheets")
	}
	rows, err := s.parseSheet(ctx, spreadsheet, sheetName)
	if err != nil {
		return model.SystemCatalogList{}, apperror.Wrap(apperror.Validation, err)
	}

	if err := s.repo.ReplaceAll(ctx, orderID, rows); err != nil {
		return model.SystemCatalogList{}, err
	}

	return s.List(ctx, model.SystemCatalogFilter{OrderID: orderID})
}

func (s *SystemCatalogService) parseSheet(ctx context.Context, spreadsheet *excelize.File, sheetName string) ([]model.SystemCatalogRow, error) {
	excelRows, err := spreadsheet.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("read system catalog sheet %q: %w", sheetName, err)
	}
	if len(excelRows) > maxImportedRows+1 {
		return nil, fmt.Errorf("sheet %q exceeds the limit of %d data rows", sheetName, maxImportedRows)
	}

	rows := make([]model.SystemCatalogRow, 0, len(excelRows))
	validationErrors := make([]string, 0)
	validationErrorCount := 0
	for index, excelRow := range excelRows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if index == 0 {
			continue
		}
		if len(excelRow) < 3 {
			if len(excelRow) > 0 && normalizeCell(strings.Join(excelRow, "")) != "" {
				addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: ожидаются шифр, название и класс", index+1))
			}
			continue
		}

		code := normalizeCell(excelRow[0])
		systemName := normalizeCell(excelRow[1])
		systemClass := normalizeStatus(excelRow[2])
		curator := ""
		if len(excelRow) > 3 {
			curator = normalizeCell(excelRow[3])
		}
		if code == "" && systemName == "" && systemClass == "" && curator == "" {
			continue
		}
		if systemName == "" || systemClass == "" {
			addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: название и класс обязательны", index+1))
			continue
		}
		if !validStatus(systemClass, false) {
			addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: недопустимый класс %q", index+1, systemClass))
			continue
		}
		if len(code) > maxImportedCellBytes || len(systemName) > maxImportedCellBytes || len(curator) > maxImportedCellBytes {
			addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: значение ячейки слишком длинное", index+1))
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
	if validationErrorCount > 0 {
		return nil, fmt.Errorf("ошибки в листе %q: %s", sheetName, formatImportValidationErrors(validationErrors, validationErrorCount))
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet %q has no system catalog rows", sheetName)
	}

	return rows, nil
}

func (s *SystemCatalogService) Update(ctx context.Context, id int64, orderID int64, row model.SystemCatalogRow) (model.SystemCatalogRow, error) {
	if id <= 0 || orderID <= 0 {
		return model.SystemCatalogRow{}, apperror.New(apperror.Validation, "invalid system catalog row")
	}
	row.Code = normalizeCell(row.Code)
	row.SystemName = normalizeCell(row.SystemName)
	row.SystemClass = normalizeStatus(row.SystemClass)
	row.Curator = normalizeCell(row.Curator)
	if row.SystemName == "" {
		return model.SystemCatalogRow{}, apperror.New(apperror.Validation, "system name cannot be empty")
	}
	if len(row.Code) > maxImportedCellBytes || len(row.SystemName) > maxImportedCellBytes || len(row.Curator) > maxImportedCellBytes {
		return model.SystemCatalogRow{}, apperror.New(apperror.Validation, "system catalog value is too long")
	}
	if !validStatus(row.SystemClass, false) {
		return model.SystemCatalogRow{}, apperror.New(apperror.Validation, "invalid system class")
	}
	return s.repo.Update(ctx, id, orderID, row)
}

func (s *SystemCatalogService) Export(ctx context.Context, filter model.SystemCatalogFilter) ([]byte, error) {
	if err := validateSystemCatalogFilter(filter); err != nil {
		return nil, err
	}
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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

func validateSystemCatalogFilter(filter model.SystemCatalogFilter) error {
	if filter.SystemClass != "" && !validStatus(filter.SystemClass, false) {
		return apperror.New(apperror.Validation, "invalid system class filter")
	}
	return nil
}
