package producers

import (
	"log"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	kafka *sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	log.Println("Kafka connected successfully!")

	return &KafkaProducer{
		kafka: &producer,
	}
}

func (kp *KafkaProducer) Produce(topic string, payload []byte) error {
	log.Println("Producing event",
		" topic", topic,
		" payload", string(payload),
	)

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := (*kp.kafka).SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)

		return err
	}

	log.Println("Message sent successfully",
		" topic", topic,
		" partition", partition,
		" offset", offset,
	)

	return nil
}
