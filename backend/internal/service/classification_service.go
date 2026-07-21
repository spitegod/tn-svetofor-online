package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"tn/backend/internal/apperror"
	"tn/backend/internal/model"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
)

type ClassificationService struct {
	repo classificationRepository
}

type classificationRepository interface {
	List(context.Context, model.ClassificationFilter) ([]model.ClassificationChange, error)
	Stats(context.Context, int64) (model.ClassificationStats, error)
	Options(context.Context, string, int64) ([]string, error)
	ReplaceAll(context.Context, int64, []model.ClassificationChange) error
	Update(context.Context, int64, int64, model.ClassificationChange) (model.ClassificationChange, error)
}

const unassignedConstructionType = "Тип не присвоен"

const (
	maxImportedRows         = 100_000
	maxImportedCellBytes    = 4_000
	maxReportedImportErrors = 50
	spreadsheetUnzipLimit   = 256 << 20
	spreadsheetXMLSizeLimit = 64 << 20
)

func spreadsheetOpenOptions() excelize.Options {
	return excelize.Options{
		UnzipSizeLimit:    spreadsheetUnzipLimit,
		UnzipXMLSizeLimit: spreadsheetXMLSizeLimit,
	}
}

type knownNAVSystem struct {
	URL              string
	ConstructionType string
}

var knownNAVSystems = map[string]knownNAVSystem{
	"тн гео полигон фрост": {
		URL:              "https://nav.tn.ru/systems/poligony-ploshchadki-khraneniya-i-pr/tn-geo-poligon-frost/",
		ConstructionType: "Специальные сооружения",
	},
	"тн гео хвостохранилище фрост": {
		URL:              "https://nav.tn.ru/systems/iskusstvennye-vodoemy-prudy-i-pr/tn-geo-khvostokhranilishche-frost/",
		ConstructionType: "Специальные сооружения",
	},
	"тн гео амбар шламовый фрост": {
		URL:              "https://nav.tn.ru/systems/iskusstvennye-vodoemy-prudy-i-pr/tn-geo-ambar-shlamovyy-frost/",
		ConstructionType: "Специальные сооружения",
	},
	"тн авиа впп фрост": {
		URL:              "https://nav.tn.ru/systems/konstruktsiya-letnogo-polya/tn-avia-vpp-frost/",
		ConstructionType: "Транспортное и дорожное строительство",
	},
	"тн кровля солид керамзит": {
		URL:              "https://nav.tn.ru/systems/ploskaya-krysha/tn-krovlya-solid-keramzit/",
		ConstructionType: "Промышленное и гражданское строительство",
	},
	"тн техизоляция камин": {
		ConstructionType: "Индивидуальное жилищное строительство",
	},
}

func NewClassificationService(repo classificationRepository) *ClassificationService {
	return &ClassificationService{repo: repo}
}

func (s *ClassificationService) List(ctx context.Context, filter model.ClassificationFilter) (model.ClassificationList, error) {
	if err := validateClassificationFilter(filter); err != nil {
		return model.ClassificationList{}, err
	}
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return model.ClassificationList{}, err
	}

	stats, err := s.repo.Stats(ctx, filter.OrderID)
	if err != nil {
		return model.ClassificationList{}, err
	}

	beforeOptions, err := s.repo.Options(ctx, "class_before", filter.OrderID)
	if err != nil {
		return model.ClassificationList{}, err
	}

	afterOptions, err := s.repo.Options(ctx, "class_after", filter.OrderID)
	if err != nil {
		return model.ClassificationList{}, err
	}

	return model.ClassificationList{
		Rows:          rows,
		Stats:         stats,
		BeforeOptions: beforeOptions,
		AfterOptions:  afterOptions,
	}, nil
}

func (s *ClassificationService) Import(ctx context.Context, orderID int64, file io.Reader) (model.ClassificationList, error) {
	if orderID <= 0 {
		orderID = 1
	}

	spreadsheet, err := excelize.OpenReader(file, spreadsheetOpenOptions())
	if err != nil {
		return model.ClassificationList{}, apperror.Wrap(apperror.Validation, fmt.Errorf("open excel file: %w", err))
	}
	defer spreadsheet.Close()

	sheetName := spreadsheet.GetSheetName(0)
	if sheetName == "" {
		return model.ClassificationList{}, apperror.New(apperror.Validation, "excel file has no sheets")
	}
	rows, err := s.parseSheet(ctx, spreadsheet, sheetName)
	if err != nil {
		return model.ClassificationList{}, apperror.Wrap(apperror.Validation, err)
	}

	if err := s.repo.ReplaceAll(ctx, orderID, rows); err != nil {
		return model.ClassificationList{}, err
	}

	return s.List(ctx, model.ClassificationFilter{OrderID: orderID})
}

func (s *ClassificationService) parseSheet(ctx context.Context, spreadsheet *excelize.File, sheetName string) ([]model.ClassificationChange, error) {
	excelRows, err := spreadsheet.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("read classification sheet %q: %w", sheetName, err)
	}
	if len(excelRows) > maxImportedRows+2 {
		return nil, fmt.Errorf("sheet %q exceeds the limit of %d data rows", sheetName, maxImportedRows)
	}

	rows := make([]model.ClassificationChange, 0, len(excelRows))
	validationErrors := make([]string, 0)
	validationErrorCount := 0
	for index, excelRow := range excelRows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if index < 2 {
			continue
		}
		if len(excelRow) < 3 {
			if len(excelRow) > 0 && normalizeCell(strings.Join(excelRow, "")) != "" {
				addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: ожидаются название, класс «было» и класс «стало»", index+1))
			}
			continue
		}

		systemName := normalizeCell(excelRow[0])
		classBefore := normalizeStatus(excelRow[1])
		classAfter := normalizeStatus(excelRow[2])
		if systemName == "" && classBefore == "" && classAfter == "" {
			continue
		}
		if systemName == "" || classBefore == "" || classAfter == "" {
			addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: не заполнены обязательные ячейки", index+1))
			continue
		}
		if !validStatus(classBefore, true) || !validStatus(classAfter, false) {
			addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: недопустимый класс %q → %q", index+1, classBefore, classAfter))
			continue
		}
		if len(systemName) > maxImportedCellBytes {
			addImportValidationError(&validationErrors, &validationErrorCount, fmt.Sprintf("строка %d: название системы слишком длинное", index+1))
			continue
		}

		row := model.ClassificationChange{
			Position:         len(rows) + 1,
			SystemName:       systemName,
			ConstructionType: unassignedConstructionType,
			ClassBefore:      classBefore,
			ClassAfter:       classAfter,
		}
		if known, found := knownNAVSystemData(systemName); found {
			row.SystemURL = known.URL
			row.ConstructionType = known.ConstructionType
		}
		rows = append(rows, row)
	}
	if validationErrorCount > 0 {
		return nil, fmt.Errorf("ошибки в листе %q: %s", sheetName, formatImportValidationErrors(validationErrors, validationErrorCount))
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet %q has no classification rows", sheetName)
	}

	return rows, nil
}

func addImportValidationError(errors *[]string, total *int, message string) {
	(*total)++
	if len(*errors) < maxReportedImportErrors {
		*errors = append(*errors, message)
	}
}

func formatImportValidationErrors(errors []string, total int) string {
	message := strings.Join(errors, "; ")
	if remaining := total - len(errors); remaining > 0 {
		message += fmt.Sprintf("; и ещё %d ошибок", remaining)
	}
	return message
}

func (s *ClassificationService) Update(ctx context.Context, id int64, orderID int64, row model.ClassificationChange) (model.ClassificationChange, error) {
	if id <= 0 || orderID <= 0 {
		return model.ClassificationChange{}, apperror.New(apperror.Validation, "invalid classification change")
	}
	row.SystemName = normalizeCell(row.SystemName)
	row.ClassBefore = normalizeStatus(row.ClassBefore)
	row.ClassAfter = normalizeStatus(row.ClassAfter)
	if row.SystemName == "" {
		return model.ClassificationChange{}, apperror.New(apperror.Validation, "system name cannot be empty")
	}
	if len(row.SystemName) > maxImportedCellBytes {
		return model.ClassificationChange{}, apperror.New(apperror.Validation, "system name is too long")
	}
	if !validStatus(row.ClassBefore, true) || !validStatus(row.ClassAfter, false) {
		return model.ClassificationChange{}, apperror.New(apperror.Validation, "invalid classification status")
	}
	return s.repo.Update(ctx, id, orderID, row)
}

func (s *ClassificationService) Export(ctx context.Context, filter model.ClassificationFilter) ([]byte, error) {
	if err := validateClassificationFilter(filter); err != nil {
		return nil, err
	}
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer file.Close()

	sheet := "Таблица 1"
	file.SetSheetName("Sheet1", sheet)
	_ = file.SetCellValue(sheet, "A1", "Название системы")
	_ = file.SetCellValue(sheet, "B1", "Класс")
	_ = file.SetCellValue(sheet, "B2", "было")
	_ = file.SetCellValue(sheet, "C2", "стало")
	_ = file.MergeCell(sheet, "A1", "A2")
	_ = file.MergeCell(sheet, "B1", "C1")

	borders := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}
	headerStyle, err := file.NewStyle(&excelize.Style{
		Border:    borders,
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9D9D9"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "1F2937"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create excel header style: %w", err)
	}
	bodyStyle, err := file.NewStyle(&excelize.Style{
		Border:    borders,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("create excel body style: %w", err)
	}
	if err := file.SetCellStyle(sheet, "A1", "C2", headerStyle); err != nil {
		return nil, fmt.Errorf("style excel header: %w", err)
	}
	_ = file.SetRowHeight(sheet, 1, 24)
	_ = file.SetRowHeight(sheet, 2, 22)

	for index, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		excelRow := index + 3
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), row.SystemName)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), row.ClassBefore)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", excelRow), row.ClassAfter)
	}
	if len(rows) > 0 {
		lastRow := len(rows) + 2
		if err := file.SetCellStyle(sheet, "A3", fmt.Sprintf("C%d", lastRow), bodyStyle); err != nil {
			return nil, fmt.Errorf("style excel rows: %w", err)
		}
		for row := 3; row <= lastRow; row++ {
			_ = file.SetRowHeight(sheet, row, 20)
		}
	}

	_ = file.SetColWidth(sheet, "A", "A", 48)
	_ = file.SetColWidth(sheet, "B", "C", 28)
	_ = file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      2,
		TopLeftCell: "A3",
		ActivePane:  "bottomLeft",
	})

	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, fmt.Errorf("write excel export: %w", err)
	}

	return buffer.Bytes(), nil
}

func validateClassificationFilter(filter model.ClassificationFilter) error {
	if filter.ClassBefore != "" && !validStatus(filter.ClassBefore, true) {
		return apperror.New(apperror.Validation, "invalid previous class filter")
	}
	if filter.ClassAfter != "" && !validStatus(filter.ClassAfter, false) {
		return apperror.New(apperror.Validation, "invalid current class filter")
	}
	if filter.ConstructionType != "" && !validConstructionType(filter.ConstructionType) {
		return apperror.New(apperror.Validation, "invalid construction type filter")
	}
	return nil
}

func normalizeCell(value string) string {
	value = strings.TrimSpace(repairMojibake(value))
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func normalizeStatus(value string) string {
	value = normalizeCell(value)
	value = strings.TrimSuffix(value, ".")

	switch strings.ToLower(value) {
	case "новая система":
		return "Новая система"
	case "рекомендованная", "рекомендованные":
		return "Рекомендованная"
	case "разрешенная", "разрешённая", "разрешенные", "разрешённые":
		return "Разрешенная"
	case "запрещенная", "запрещённая", "запрещенные", "запрещённые":
		return "Запрещенная"
	default:
		return value
	}
}

func validStatus(value string, allowNew bool) bool {
	if allowNew && value == "Новая система" {
		return true
	}
	return value == "Рекомендованная" || value == "Разрешенная" || value == "Запрещенная"
}

func validConstructionType(value string) bool {
	switch value {
	case "Промышленное и гражданское строительство",
		"Индивидуальное жилищное строительство",
		"Транспортное и дорожное строительство",
		"Специальные сооружения",
		unassignedConstructionType:
		return true
	default:
		return false
	}
}

func repairMojibake(value string) string {
	if !looksLikeMojibake(value) {
		return value
	}

	encoded, err := charmap.Windows1251.NewEncoder().String(value)
	if err != nil {
		return value
	}

	if utf8.ValidString(encoded) {
		return encoded
	}

	return value
}

func looksLikeMojibake(value string) bool {
	markers := []string{"Рќ", "Рў", "РЎ", "Рџ", "Р ", "Рљ", "Р‘", "Р°", "Рµ", "Рё", "СЃ", "С‚", "СЏ", "СЊ"}
	for _, marker := range markers {
		if strings.Contains(value, marker) {
			return true
		}
	}

	return false
}

func knownNAVSystemData(systemName string) (knownNAVSystem, bool) {
	system, found := knownNAVSystems[normalizeNAVLookupName(systemName)]
	return system, found
}

func normalizeNAVLookupName(value string) string {
	normalized := normalizeSystemName(value)
	return strings.TrimSpace(strings.TrimPrefix(normalized, "система "))
}

func normalizeConstructionType(value string) string {
	normalized := normalizeSystemName(value)
	switch {
	case normalized == "пгс" || strings.Contains(normalized, "промышлен") || strings.Contains(normalized, "гражданск"):
		return "Промышленное и гражданское строительство"
	case normalized == "ижс" || strings.Contains(normalized, "индивидуальн") || strings.Contains(normalized, "жилищн"):
		return "Индивидуальное жилищное строительство"
	case strings.Contains(normalized, "транспорт") || strings.Contains(normalized, "дорож"):
		return "Транспортное и дорожное строительство"
	case normalized == "сс" || strings.Contains(normalized, "специальн") || strings.Contains(normalized, "спецсооруж") || strings.Contains(normalized, "спец сооруж"):
		return "Специальные сооружения"
	default:
		return unassignedConstructionType
	}
}
