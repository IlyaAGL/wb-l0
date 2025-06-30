package interfaces

import "github.com/agl/wbtech/internal/domain/entities"

type OrderRepository interface {
	GetOrderByID(id string) (*entities.Order, error)
	StoreOrder(mshChan chan []byte) error
}
