package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"tn/backend/internal/model"
	"tn/backend/internal/repository"

	"github.com/xuri/excelize/v2"
)

type OrderService struct {
	repo           *repository.OrderRepository
	classification *ClassificationService
	systemCatalog  *SystemCatalogService
}

func NewOrderService(repo *repository.OrderRepository, classification *ClassificationService, systemCatalog *SystemCatalogService) *OrderService {
	return &OrderService{repo: repo, classification: classification, systemCatalog: systemCatalog}
}

func (s *OrderService) List(ctx context.Context) ([]model.Order, error) {
	return s.repo.List(ctx)
}

func (s *OrderService) Create(ctx context.Context, name string) (model.Order, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Order{}, fmt.Errorf("order name is required")
	}

	return s.repo.Create(ctx, name)
}

func (s *OrderService) ImportWorkbook(ctx context.Context, name string, file io.Reader) (model.Order, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Order{}, fmt.Errorf("order name is required")
	}

	spreadsheet, err := excelize.OpenReader(file)
	if err != nil {
		return model.Order{}, fmt.Errorf("open order workbook: %w", err)
	}
	defer spreadsheet.Close()

	sheets := spreadsheet.GetSheetList()
	if len(sheets) < 2 {
		return model.Order{}, fmt.Errorf("order workbook must contain at least two sheets")
	}

	classificationRows, err := s.classification.parseSheet(ctx, spreadsheet, sheets[0])
	if err != nil {
		return model.Order{}, fmt.Errorf("import table 1: %w", err)
	}
	systemCatalogRows, err := s.systemCatalog.parseSheet(spreadsheet, sheets[1])
	if err != nil {
		return model.Order{}, fmt.Errorf("import table 2: %w", err)
	}

	order, err := s.repo.Create(ctx, name)
	if err != nil {
		return model.Order{}, err
	}

	if err := s.classification.repo.ReplaceAll(ctx, order.ID, classificationRows); err != nil {
		_ = s.repo.Delete(ctx, order.ID)
		return model.Order{}, fmt.Errorf("save table 1: %w", err)
	}
	if err := s.systemCatalog.repo.ReplaceAll(ctx, order.ID, systemCatalogRows); err != nil {
		_ = s.repo.Delete(ctx, order.ID)
		return model.Order{}, fmt.Errorf("save table 2: %w", err)
	}

	return order, nil
}

func (s *OrderService) UpdateName(ctx context.Context, id int64, name string) (model.Order, error) {
	if id <= 0 {
		return model.Order{}, fmt.Errorf("invalid order id")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return model.Order{}, fmt.Errorf("order name is required")
	}

	return s.repo.UpdateName(ctx, id, name)
}

func (s *OrderService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid order id")
	}

	return s.repo.Delete(ctx, id)
}
