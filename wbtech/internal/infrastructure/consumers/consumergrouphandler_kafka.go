package consumers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/agl/wbtech/internal/domain/entities"
	"github.com/agl/wbtech/pkg/logger"
)

type ConsumerGroupHandler struct {
	msgChan chan<- []byte
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Log.Info("Received message", "value", string(msg.Value))
		var event entities.Order
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Log.Error("Error unmarshalling message", "error", err)
			continue
		}
		if err := validateOrder(&event); err != nil {
			logger.Log.Error("Order validation failed", "error", err, "order_uid", event.OrderUID)
			continue
		}
		logger.Log.Info("Message is correctly unmarshalled", "order_uid", event.OrderUID)
		h.msgChan <- msg.Value
		session.MarkMessage(msg, "")
	}
	return nil
}

func validateOrder(o *entities.Order) error {
	if o.OrderUID == "" || o.TrackNumber == "" || o.Entry == "" ||
		o.Locale == "" || o.CustomerID == "" || o.DeliveryService == "" ||
		o.ShardKey == "" || o.SmID == 0 || o.DateCreated == "" || o.OofShard == "" {
		return errors.New("missing required order fields")
	}
	if o.Delivery.Name == "" || o.Delivery.Phone == "" || o.Delivery.Zip == "" ||
		o.Delivery.City == "" || o.Delivery.Address == "" || o.Delivery.Region == "" || o.Delivery.Email == "" {
		return errors.New("missing required delivery fields")
	}
	if o.Payment.Transaction == "" || o.Payment.Currency == "" || o.Payment.Provider == "" ||
		o.Payment.Amount == 0 || o.Payment.PaymentDT == 0 || o.Payment.Bank == "" {
		return errors.New("missing required payment fields")
	}
	if len(o.Items) == 0 {
		return errors.New("items is empty")
	}
	for i, item := range o.Items {
		if item.ChrtID == 0 || item.TrackNumber == "" || item.Price == 0 ||
			item.Rid == "" || item.Name == "" || item.Size == "" ||
			item.TotalPrice == 0 || item.NmID == 0 || item.Brand == "" || item.Status == 0 {
			return fmt.Errorf("missing required item fields in item %d", i)
		}
	}
	return nil
}
