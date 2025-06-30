package services

import (
	"encoding/json"

	"github.com/agl/wbtech/internal/application/dto"
	"github.com/agl/wbtech/internal/application/interfaces"
	"github.com/agl/wbtech/internal/domain/entities"
	"github.com/agl/wbtech/pkg/logger"
)

type OrderService struct {
	repo interfaces.OrderRepository
}

func NewOrderService(repo interfaces.OrderRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (s *OrderService) GetOrderByID(id string) (*dto.Order, error) {
	order, err := s.repo.GetOrderByID(id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, nil
	}
	return ConvertOrderToDTO(order)
}

func (s *OrderService) StoreOrder(msgChan chan []byte) error {
	if err := s.repo.StoreOrder(msgChan); err != nil {
		logger.Log.Error("Failed to store order", "error", err)

		return err
	}

	return nil
}

func ConvertOrderToDTO(o *entities.Order) (*dto.Order, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	var dtoOrder dto.Order
	if err := json.Unmarshal(b, &dtoOrder); err != nil {
		return nil, err
	}
	return &dtoOrder, nil
}

func ConvertDTOToEntity(d *dto.Order) (*entities.Order, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	var entity entities.Order
	if err := json.Unmarshal(b, &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}
