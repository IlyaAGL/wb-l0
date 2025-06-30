package consumers

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/agl/wbtech/pkg/logger"
)

const topic = "service.message"

type KafkaConsumer struct {
	Kafka sarama.ConsumerGroup
}

func NewKafkaConsumer(brokers []string, groupID string) *KafkaConsumer {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_1_0_0

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		logger.Log.Error("Couldn't create kafka consumer group", "error", err)
		return nil
	}

	logger.Log.Info("Kafka consumer group created successfully")

	return &KafkaConsumer{
		Kafka: consumerGroup,
	}
}

func (kc *KafkaConsumer) Consume(msgChan chan<- []byte) {
	handler := &ConsumerGroupHandler{msgChan: msgChan}
	ctx := context.Background()
	go func() {
		for {
			if err := kc.Kafka.Consume(ctx, []string{topic}, handler); err != nil {
				logger.Log.Error("Error from consumer group", "error", err)
			}
		}
	}()
}
