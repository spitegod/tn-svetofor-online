package service

import (
	"context"
	"fmt"
	"strings"

	"tn/backend/internal/model"
	"tn/backend/internal/repository"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
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
