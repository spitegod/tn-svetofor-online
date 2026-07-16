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
	file := excelize.NewFile()
	defer file.Close()
	sheet := "Список систем"
	file.SetSheetName("Sheet1", sheet)
	headers := []string{"Шифр", "Название системы", "Класс", "Куратор", "Комментарий", "Документ"}
	for index, header := range headers {
		_ = file.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+index), header)
	}
	for index, row := range rows {
		excelRow := index + 2
		values := []string{row.Code, row.SystemName, row.SystemClass, row.Curator, row.Comment, row.AttachmentName}
		for column, value := range values {
			_ = file.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+column, excelRow), value)
		}
	}
	_ = file.SetColWidth(sheet, "A", "A", 18)
	_ = file.SetColWidth(sheet, "B", "B", 48)
	_ = file.SetColWidth(sheet, "C", "F", 24)
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, fmt.Errorf("write system document export: %w", err)
	}
	return buffer.Bytes(), nil
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
