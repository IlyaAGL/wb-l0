package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/agl/emulator/entities"
	"github.com/agl/emulator/producers"
	"github.com/gin-gonic/gin"
)

const topic = "service.message"

type Server struct {
	kafka *producers.KafkaProducer
}

func NewServer(kafka *producers.KafkaProducer) *Server {
	return &Server{
		kafka: kafka,
	}
}

func (s *Server) StartServer() {
	r := gin.Default()

	r.POST("/produce", s.produce)

	if err := r.Run(":1234"); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) produce(ctx *gin.Context) {
    var order entities.Order
    if err := ctx.ShouldBindJSON(&order); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid input",
            "details": err.Error(),
        })
        return
    }

    orderBytes, err := json.Marshal(order)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Failed to serialize order",
            "details": err.Error(),
        })
        return
    }

    if err := s.kafka.Produce(topic, orderBytes); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Failed to produce message",
            "details": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
