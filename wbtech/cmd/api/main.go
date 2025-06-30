package main

import (
	"github.com/agl/wbtech/internal/application/handlers"
	"github.com/agl/wbtech/internal/application/services"
	"github.com/agl/wbtech/internal/infrastructure/consumers"
	"github.com/agl/wbtech/internal/infrastructure/repositories"
	"github.com/agl/wbtech/internal/presentation/controllers"
	"github.com/agl/wbtech/pkg/dbconnections"
)

var brokers = []string{
	"kafka:9092",
}

const groupID = "order-api-group"

func main() {
	db_pg := dbconnections.InitPostgres()
	defer db_pg.Close()

	repo := repositories.NewOrderRepository(db_pg)
	service := services.NewOrderService(repo)
	controller := controllers.NewOrderController(service)

	consumer := consumers.NewKafkaConsumer(brokers, groupID)
	msg_handler := handlers.NewMessageHandler(consumer, service)

	go msg_handler.HandleMessage()

	controller.StartServer()
}
