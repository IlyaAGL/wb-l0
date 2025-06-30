package main

import (
	"github.com/agl/emulator/producers"
	"github.com/agl/emulator/server"
)

var brokers = []string {
	"kafka:9092",
}

func main() {
	producer := producers.NewKafkaProducer(brokers)
	server := server.NewServer(producer)

	server.StartServer()
}
