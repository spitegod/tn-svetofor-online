package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"tn/backend/internal/apperror"
	"tn/backend/internal/model"

	"github.com/xuri/excelize/v2"
)

type OrderService struct {
	repo           orderRepository
	classification *ClassificationService
	systemCatalog  *SystemCatalogService
}

type orderRepository interface {
	List(context.Context) ([]model.Order, error)
	Create(context.Context, string) (model.Order, error)
	Import(context.Context, string, []model.ClassificationChange, []model.SystemCatalogRow) (model.Order, error)
	UpdateName(context.Context, int64, string) (model.Order, error)
	Delete(context.Context, int64) error
}

const maxOrderNameBytes = 500

func NewOrderService(repo orderRepository, classification *ClassificationService, systemCatalog *SystemCatalogService) *OrderService {
	return &OrderService{repo: repo, classification: classification, systemCatalog: systemCatalog}
}

func (s *OrderService) List(ctx context.Context) ([]model.Order, error) {
	return s.repo.List(ctx)
}

func (s *OrderService) Create(ctx context.Context, name string) (model.Order, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Order{}, apperror.New(apperror.Validation, "order name is required")
	}
	if len(name) > maxOrderNameBytes {
		return model.Order{}, apperror.New(apperror.Validation, "order name is too long")
	}

	return s.repo.Create(ctx, name)
}

func (s *OrderService) ImportWorkbook(ctx context.Context, name string, file io.Reader) (model.Order, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Order{}, apperror.New(apperror.Validation, "order name is required")
	}
	if len(name) > maxOrderNameBytes {
		return model.Order{}, apperror.New(apperror.Validation, "order name is too long")
	}

	spreadsheet, err := excelize.OpenReader(file, spreadsheetOpenOptions())
	if err != nil {
		return model.Order{}, apperror.Wrap(apperror.Validation, fmt.Errorf("open order workbook: %w", err))
	}
	defer spreadsheet.Close()

	sheets := spreadsheet.GetSheetList()
	if len(sheets) < 2 {
		return model.Order{}, apperror.New(apperror.Validation, "order workbook must contain at least two sheets")
	}

	classificationRows, err := s.classification.parseSheet(ctx, spreadsheet, sheets[0])
	if err != nil {
		return model.Order{}, apperror.Wrap(apperror.Validation, fmt.Errorf("import table 1: %w", err))
	}
	systemCatalogRows, err := s.systemCatalog.parseSheet(ctx, spreadsheet, sheets[1])
	if err != nil {
		return model.Order{}, apperror.Wrap(apperror.Validation, fmt.Errorf("import table 2: %w", err))
	}

	return s.repo.Import(ctx, name, classificationRows, systemCatalogRows)
}

func (s *OrderService) UpdateName(ctx context.Context, id int64, name string) (model.Order, error) {
	if id <= 0 {
		return model.Order{}, apperror.New(apperror.Validation, "invalid order id")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return model.Order{}, apperror.New(apperror.Validation, "order name is required")
	}
	if len(name) > maxOrderNameBytes {
		return model.Order{}, apperror.New(apperror.Validation, "order name is too long")
	}

	return s.repo.UpdateName(ctx, id, name)
}

func (s *OrderService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return apperror.New(apperror.Validation, "invalid order id")
	}

	return s.repo.Delete(ctx, id)
}
