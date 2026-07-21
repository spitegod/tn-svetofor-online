package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"tn/backend/internal/apperror"
	"tn/backend/internal/model"

	"github.com/xuri/excelize/v2"
)

type SystemDocumentService struct {
	repo systemDocumentRepository
}

type systemDocumentRepository interface {
	List(context.Context, model.SystemDocumentFilter) ([]model.SystemDocumentRow, error)
	Stats(context.Context, int64) (model.SystemCatalogStats, error)
	Options(context.Context, string, int64) ([]string, error)
	History(context.Context, string, string) ([]model.SystemDocumentRow, error)
	UpdateComment(context.Context, int64, int64, string) (model.SystemDocumentRow, error)
	SaveAttachment(context.Context, int64, int64, model.SystemDocumentAttachment) error
	Attachment(context.Context, int64, int64) (model.SystemDocumentAttachment, error)
	DeleteAttachment(context.Context, int64, int64) error
	UpdateComparison(context.Context, int64, int64, bool) error
	UpdateComparisonBulk(context.Context, int64, bool, bool, []model.SystemDocumentKey) error
}

const MaxSystemDocumentAttachmentSize int64 = 25 << 20

const maxSystemDocumentCommentBytes = 20_000

func (s *SystemDocumentService) Export(ctx context.Context, filter model.SystemDocumentFilter) ([]byte, error) {
	if err := validateSystemDocumentFilter(filter); err != nil {
		return nil, err
	}
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	rows = filterSystemDocumentExportRows(rows, filter)

	file := excelize.NewFile()
	defer file.Close()
	sheet := "Список систем"
	if err := file.SetSheetName("Sheet1", sheet); err != nil {
		return nil, fmt.Errorf("rename system document export sheet: %w", err)
	}
	headers := []string{"Шифр", "Название системы", "Класс", "Куратор", "Комментарий", "Документ"}
	for index, header := range headers {
		_ = file.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+index), header)
	}

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
		return nil, fmt.Errorf("create system document excel header style: %w", err)
	}
	bodyStyle, err := file.NewStyle(&excelize.Style{
		Border:    borders,
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create system document excel body style: %w", err)
	}
	if err := file.SetCellStyle(sheet, "A1", "F1", headerStyle); err != nil {
		return nil, fmt.Errorf("style system document excel header: %w", err)
	}
	_ = file.SetRowHeight(sheet, 1, 24)

	for index, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		excelRow := index + 2
		values := []string{row.Code, row.SystemName, row.SystemClass, row.Curator, row.Comment, row.AttachmentName}
		for column, value := range values {
			_ = file.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+column, excelRow), value)
		}
	}
	if len(rows) > 0 {
		lastRow := len(rows) + 1
		if err := file.SetCellStyle(sheet, "A2", fmt.Sprintf("F%d", lastRow), bodyStyle); err != nil {
			return nil, fmt.Errorf("style system document excel rows: %w", err)
		}
		for row := 2; row <= lastRow; row++ {
			_ = file.SetRowHeight(sheet, row, 20)
		}
	}
	_ = file.SetColWidth(sheet, "A", "A", 18)
	_ = file.SetColWidth(sheet, "B", "B", 48)
	_ = file.SetColWidth(sheet, "C", "D", 24)
	_ = file.SetColWidth(sheet, "E", "E", 42)
	_ = file.SetColWidth(sheet, "F", "F", 28)
	_ = file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, fmt.Errorf("write system document export: %w", err)
	}
	return buffer.Bytes(), nil
}

func (s *SystemDocumentService) ExportComparison(ctx context.Context, payload model.ComparisonExport) ([]byte, error) {
	if len(payload.Headers) < 2 {
		return nil, apperror.New(apperror.Validation, "comparison export has no columns")
	}
	if len(payload.Headers) > 20 || len(payload.Rows) > 50_000 {
		return nil, apperror.New(apperror.Validation, "comparison export is too large")
	}
	for _, header := range payload.Headers {
		if len(header) > maxImportedCellBytes {
			return nil, apperror.New(apperror.Validation, "comparison header is too long")
		}
	}
	for _, row := range payload.Rows {
		for _, value := range row {
			if len(value) > maxImportedCellBytes {
				return nil, apperror.New(apperror.Validation, "comparison cell is too long")
			}
		}
	}

	file := excelize.NewFile()
	defer file.Close()
	sheet := "Сравнение"
	if err := file.SetSheetName("Sheet1", sheet); err != nil {
		return nil, fmt.Errorf("rename comparison export sheet: %w", err)
	}

	borders := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}
	headerStyle, err := file.NewStyle(&excelize.Style{
		Border: borders,
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"D9D9D9"}, Pattern: 1},
		Font:   &excelize.Font{Bold: true, Color: "1F2937"},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create comparison header style: %w", err)
	}
	bodyStyle, err := file.NewStyle(&excelize.Style{
		Border: borders,
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create comparison body style: %w", err)
	}

	for column, value := range payload.Headers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		cell, _ := excelize.CoordinatesToCellName(column+1, 1)
		_ = file.SetCellValue(sheet, cell, value)
	}
	for rowIndex, row := range payload.Rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for column := range payload.Headers {
			value := ""
			if column < len(row) {
				value = row[column]
			}
			cell, _ := excelize.CoordinatesToCellName(column+1, rowIndex+2)
			_ = file.SetCellValue(sheet, cell, value)
		}
	}

	lastColumn, _ := excelize.ColumnNumberToName(len(payload.Headers))
	if err := file.SetCellStyle(sheet, "A1", lastColumn+"1", headerStyle); err != nil {
		return nil, fmt.Errorf("style comparison header: %w", err)
	}
	if len(payload.Rows) > 0 {
		lastRow := len(payload.Rows) + 1
		if err := file.SetCellStyle(sheet, "A2", fmt.Sprintf("%s%d", lastColumn, lastRow), bodyStyle); err != nil {
			return nil, fmt.Errorf("style comparison rows: %w", err)
		}
		for row := 2; row <= lastRow; row++ {
			_ = file.SetRowHeight(sheet, row, 20)
		}
	}
	_ = file.SetRowHeight(sheet, 1, 26)
	_ = file.SetColWidth(sheet, "A", "A", 48)
	if len(payload.Headers) > 1 {
		_ = file.SetColWidth(sheet, "B", lastColumn, 28)
	}
	_ = file.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, fmt.Errorf("write comparison export: %w", err)
	}
	return buffer.Bytes(), nil
}

func filterSystemDocumentExportRows(rows []model.SystemDocumentRow, filter model.SystemDocumentFilter) []model.SystemDocumentRow {
	if filter.ConstructionType == "" && filter.SystemType == "" {
		return rows
	}

	filtered := make([]model.SystemDocumentRow, 0, len(rows))
	for _, row := range rows {
		if filter.ConstructionType != "" && !hasSystemCharacteristic(row, "Сегмент строительства", filter.ConstructionType, true) {
			continue
		}
		if filter.SystemType != "" && !hasSystemCharacteristic(row, "Тип системы", filter.SystemType, false) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func hasSystemCharacteristic(row model.SystemDocumentRow, name string, value string, contains bool) bool {
	for _, characteristic := range row.Characteristics {
		if characteristic.Name != name {
			continue
		}
		if contains && strings.Contains(characteristic.Value, value) {
			return true
		}
		if !contains && characteristic.Value == value {
			return true
		}
	}
	return false
}

func NewSystemDocumentService(repo systemDocumentRepository) *SystemDocumentService {
	return &SystemDocumentService{repo: repo}
}

func (s *SystemDocumentService) List(ctx context.Context, filter model.SystemDocumentFilter) (model.SystemDocumentList, error) {
	if filter.OrderID <= 0 {
		filter.OrderID = 1
	}
	if err := validateSystemDocumentFilter(filter); err != nil {
		return model.SystemDocumentList{}, err
	}
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return model.SystemDocumentList{}, err
	}
	stats, err := s.repo.Stats(ctx, filter.OrderID)
	if err != nil {
		return model.SystemDocumentList{}, err
	}
	classOptions, err := s.repo.Options(ctx, "system_class", filter.OrderID)
	if err != nil {
		return model.SystemDocumentList{}, err
	}
	curatorOptions, err := s.repo.Options(ctx, "curator", filter.OrderID)
	if err != nil {
		return model.SystemDocumentList{}, err
	}
	return model.SystemDocumentList{Rows: rows, Stats: stats, ClassOptions: classOptions, CuratorOptions: curatorOptions}, nil
}

func validateSystemDocumentFilter(filter model.SystemDocumentFilter) error {
	if filter.SystemClass != "" && !validStatus(filter.SystemClass, false) {
		return apperror.New(apperror.Validation, "invalid system class filter")
	}
	if filter.ConstructionType != "" && !validConstructionType(filter.ConstructionType) {
		return apperror.New(apperror.Validation, "invalid construction type filter")
	}
	return nil
}

func (s *SystemDocumentService) History(ctx context.Context, code string, systemName string) ([]model.SystemDocumentRow, error) {
	code = strings.TrimSpace(code)
	systemName = strings.TrimSpace(systemName)
	if code == "" && systemName == "" {
		return nil, apperror.New(apperror.Validation, "system code or name is required")
	}
	if len(code) > maxImportedCellBytes || len(systemName) > maxImportedCellBytes {
		return nil, apperror.New(apperror.Validation, "system code or name is too long")
	}
	return s.repo.History(ctx, code, systemName)
}

func (s *SystemDocumentService) UpdateComment(ctx context.Context, id int64, orderID int64, comment string) (model.SystemDocumentRow, error) {
	if id <= 0 || orderID <= 0 {
		return model.SystemDocumentRow{}, apperror.New(apperror.Validation, "invalid system document id or order id")
	}
	if len(comment) > maxSystemDocumentCommentBytes {
		return model.SystemDocumentRow{}, apperror.New(apperror.Validation, "comment exceeds %d bytes", maxSystemDocumentCommentBytes)
	}
	return s.repo.UpdateComment(ctx, id, orderID, comment)
}

func (s *SystemDocumentService) SaveAttachment(ctx context.Context, id int64, orderID int64, attachment model.SystemDocumentAttachment) error {
	if id <= 0 || orderID <= 0 {
		return apperror.New(apperror.Validation, "invalid system document id or order id")
	}
	attachment.Name = filepath.Base(strings.ReplaceAll(strings.TrimSpace(attachment.Name), "\\", "/"))
	attachment.Size = int64(len(attachment.Data))
	if attachment.Name == "" || attachment.Name == "." || attachment.Size == 0 {
		return apperror.New(apperror.Validation, "attachment file is empty")
	}
	if len(attachment.Name) > 255 {
		return apperror.New(apperror.Validation, "attachment file name is too long")
	}
	if attachment.Size > MaxSystemDocumentAttachmentSize {
		return apperror.New(apperror.Validation, "attachment exceeds 25 MB")
	}
	extension := strings.ToLower(filepath.Ext(attachment.Name))
	if extension != ".pdf" && extension != ".doc" && extension != ".docx" {
		return apperror.New(apperror.Validation, "attachment must be PDF, DOC or DOCX")
	}
	contentType, err := validateSystemDocumentAttachment(extension, attachment.Data)
	if err != nil {
		return apperror.Wrap(apperror.Validation, err)
	}
	attachment.ContentType = contentType
	return s.repo.SaveAttachment(ctx, id, orderID, attachment)
}

func validateSystemDocumentAttachment(extension string, data []byte) (string, error) {
	switch extension {
	case ".pdf":
		if !bytes.HasPrefix(data, []byte("%PDF-")) {
			return "", fmt.Errorf("attachment content does not match PDF format")
		}
		return "application/pdf", nil
	case ".doc":
		legacyWordSignature := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
		if !bytes.HasPrefix(data, legacyWordSignature) {
			return "", fmt.Errorf("attachment content does not match DOC format")
		}
		return "application/msword", nil
	case ".docx":
		archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return "", fmt.Errorf("attachment content does not match DOCX format")
		}
		hasContentTypes := false
		hasWordDocument := false
		for _, file := range archive.File {
			switch file.Name {
			case "[Content_Types].xml":
				hasContentTypes = true
			case "word/document.xml":
				hasWordDocument = true
			}
		}
		if !hasContentTypes || !hasWordDocument {
			return "", fmt.Errorf("attachment content does not match DOCX format")
		}
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil
	default:
		return "", fmt.Errorf("unsupported attachment format")
	}
}

func (s *SystemDocumentService) Attachment(ctx context.Context, id int64, orderID int64) (model.SystemDocumentAttachment, error) {
	if id <= 0 || orderID <= 0 {
		return model.SystemDocumentAttachment{}, apperror.New(apperror.Validation, "invalid system document id or order id")
	}
	return s.repo.Attachment(ctx, id, orderID)
}

func (s *SystemDocumentService) DeleteAttachment(ctx context.Context, id int64, orderID int64) error {
	if id <= 0 || orderID <= 0 {
		return apperror.New(apperror.Validation, "invalid system document id or order id")
	}
	return s.repo.DeleteAttachment(ctx, id, orderID)
}

func (s *SystemDocumentService) UpdateComparison(ctx context.Context, id int64, orderID int64, selected bool) error {
	if id <= 0 || orderID <= 0 {
		return apperror.New(apperror.Validation, "invalid system document id or order id")
	}
	return s.repo.UpdateComparison(ctx, id, orderID, selected)
}

func (s *SystemDocumentService) UpdateComparisonBulk(ctx context.Context, orderID int64, allOrders bool, selected bool, systems []model.SystemDocumentKey) error {
	if orderID <= 0 {
		return apperror.New(apperror.Validation, "invalid order id")
	}
	if len(systems) > 5_000 {
		return apperror.New(apperror.Validation, "too many systems in one comparison update")
	}
	return s.repo.UpdateComparisonBulk(ctx, orderID, allOrders, selected, systems)
}
