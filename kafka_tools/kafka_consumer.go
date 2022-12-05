package kafka_tools

import (
	"context"
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

type kafkaConsumer struct {
	ListenTopics  []string
	ConsumerGroup sarama.ConsumerGroup
}

func (consumer *kafkaConsumer) ConsumerMsgAndBlock(ctx context.Context, f ConsumeFunc) error {
	handler := NewConsumerGroupHandler(f)

	for {
		err := consumer.ConsumerGroup.Consume(ctx, consumer.ListenTopics, handler)
		if err != nil {
			log.Error("program quit with error: ", err.Error())
			return err
		}
	}
}
