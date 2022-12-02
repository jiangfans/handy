package kafka_tools

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

type kafkaConsumer struct {
	ConsumerGroup sarama.ConsumerGroup
}

func (consumer *kafkaConsumer) ConsumerMsgAndBlock(ctx context.Context, f ConsumeFunc) {

}

func SendMessage(topic string, message []byte) error {
	producer, err := sarama.NewSyncProducer(config.Cfg.KafkaHost, nil)
	if err != nil {
		return err
	}
	defer func() {
		producer.Close()
	}()

	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(message)}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("FAILED to send message: %s, topic:%s\n", err, topic)
	} else {
		log.Printf("> message sent to partition %d at offset %d, topic:%s\n", partition, offset, topic)
	}
	return err
}

type GroupConsumer struct {
	ProcessFunc func(*sarama.ConsumerMessage) error
	KafkaServer []string
	ListenTopic []string
	GroupId     string
}

func (h *GroupConsumer) Run(ctx context.Context) error {
	if h.GroupId == "" || len(h.ListenTopic) < 1 || len(h.KafkaServer) < 1 {
		panic(fmt.Sprintf("kafka config invalid, GroupId:%s ListenTopic:%s KafkaServer:%v\n", h.GroupId, h.ListenTopic, h.KafkaServer))
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConfig.Version = sarama.V1_1_0_0

	client, err := sarama.NewClient(h.KafkaServer, saramaConfig)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	group, err := sarama.NewConsumerGroupFromClient(h.GroupId, client)
	if err != nil {
		return err
	}

	for {
		err = group.Consume(ctx, h.ListenTopic, h)
		if err != nil {
			return err
		}
	}

	return nil
}
