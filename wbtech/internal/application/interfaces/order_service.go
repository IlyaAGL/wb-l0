package interfaces

import "github.com/agl/wbtech/internal/application/dto"

type OrderService interface {
	GetOrderByID(id string) (*dto.Order, error)
	StoreOrder(msgChan chan []byte) error
}
