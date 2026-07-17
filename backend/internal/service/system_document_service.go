package service

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"tn/backend/internal/model"
	"tn/backend/internal/repository"

	"github.com/xuri/excelize/v2"
)

type SystemDocumentService struct {
	repo *repository.SystemDocumentRepository
}

const MaxSystemDocumentAttachmentSize int64 = 25 << 20

func (s *SystemDocumentService) Export(ctx context.Context, filter model.SystemDocumentFilter) ([]byte, error) {
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

func (s *SystemDocumentService) ExportComparison(payload model.ComparisonExport) ([]byte, error) {
	if len(payload.Headers) < 2 {
		return nil, fmt.Errorf("comparison export has no columns")
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
		cell, _ := excelize.CoordinatesToCellName(column+1, 1)
		_ = file.SetCellValue(sheet, cell, value)
	}
	for rowIndex, row := range payload.Rows {
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

func NewSystemDocumentService(repo *repository.SystemDocumentRepository) *SystemDocumentService {
	return &SystemDocumentService{repo: repo}
}

func (s *SystemDocumentService) List(ctx context.Context, filter model.SystemDocumentFilter) (model.SystemDocumentList, error) {
	if filter.OrderID <= 0 {
		filter.OrderID = 1
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

func (s *SystemDocumentService) History(ctx context.Context, code string, systemName string) ([]model.SystemDocumentRow, error) {
	if code == "" || systemName == "" {
		return nil, fmt.Errorf("system code and name are required")
	}
	return s.repo.History(ctx, code, systemName)
}

func (s *SystemDocumentService) UpdateComment(ctx context.Context, id int64, orderID int64, comment string) (model.SystemDocumentRow, error) {
	if id <= 0 || orderID <= 0 {
		return model.SystemDocumentRow{}, fmt.Errorf("invalid system document id or order id")
	}
	return s.repo.UpdateComment(ctx, id, orderID, comment)
}

func (s *SystemDocumentService) SaveAttachment(ctx context.Context, id int64, orderID int64, attachment model.SystemDocumentAttachment) error {
	if id <= 0 || orderID <= 0 {
		return fmt.Errorf("invalid system document id or order id")
	}
	attachment.Name = filepath.Base(strings.TrimSpace(attachment.Name))
	attachment.Size = int64(len(attachment.Data))
	if attachment.Name == "" || attachment.Name == "." || attachment.Size == 0 {
		return fmt.Errorf("attachment file is empty")
	}
	if attachment.Size > MaxSystemDocumentAttachmentSize {
		return fmt.Errorf("attachment exceeds 25 MB")
	}
	extension := strings.ToLower(filepath.Ext(attachment.Name))
	if extension != ".pdf" && extension != ".doc" && extension != ".docx" {
		return fmt.Errorf("attachment must be PDF, DOC or DOCX")
	}
	if attachment.ContentType == "" {
		attachment.ContentType = "application/octet-stream"
	}
	return s.repo.SaveAttachment(ctx, id, orderID, attachment)
}

func (s *SystemDocumentService) Attachment(ctx context.Context, id int64, orderID int64) (model.SystemDocumentAttachment, error) {
	if id <= 0 || orderID <= 0 {
		return model.SystemDocumentAttachment{}, fmt.Errorf("invalid system document id or order id")
	}
	return s.repo.Attachment(ctx, id, orderID)
}

func (s *SystemDocumentService) DeleteAttachment(ctx context.Context, id int64, orderID int64) error {
	if id <= 0 || orderID <= 0 {
		return fmt.Errorf("invalid system document id or order id")
	}
	return s.repo.DeleteAttachment(ctx, id, orderID)
}

func (s *SystemDocumentService) Delete(ctx context.Context, id int64, orderID int64) error {
	if id <= 0 || orderID <= 0 {
		return fmt.Errorf("invalid system document id or order id")
	}
	return s.repo.Delete(ctx, id, orderID)
}

func (s *SystemDocumentService) UpdateComparison(ctx context.Context, id int64, orderID int64, selected bool) error {
	if id <= 0 || orderID <= 0 {
		return fmt.Errorf("invalid system document id or order id")
	}
	return s.repo.UpdateComparison(ctx, id, orderID, selected)
}

func (s *SystemDocumentService) UpdateComparisonBulk(ctx context.Context, orderID int64, allOrders bool, selected bool, systems []model.SystemDocumentKey) error {
	if orderID <= 0 {
		return fmt.Errorf("invalid order id")
	}
	return s.repo.UpdateComparisonBulk(ctx, orderID, allOrders, selected, systems)
}
