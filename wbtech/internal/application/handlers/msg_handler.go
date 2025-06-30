package handlers

import (
	"github.com/agl/wbtech/internal/application/interfaces"
	"github.com/agl/wbtech/pkg/logger"
)

type MessageHandler struct {
	consumer interfaces.Consumer
	service  interfaces.OrderService
}

func NewMessageHandler(consumer interfaces.Consumer, service interfaces.OrderService) *MessageHandler {
	return &MessageHandler{
		consumer: consumer,
		service:  service,
	}
}

func (mh *MessageHandler) HandleMessage() {
	msgChan := make(chan []byte)

	mh.consumer.Consume(msgChan)

	logger.Log.Info("Start storing the message")

	if err := mh.service.StoreOrder(msgChan); err != nil {
		logger.Log.Error("Something went wrong during storing the message", "error", err)

		return
	}

	logger.Log.Info("Message stored successfully :)")
}
