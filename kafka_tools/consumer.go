package kafka_tools

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

const (
	OffsetNewest = sarama.OffsetNewest
	OffsetOldest = sarama.OffsetOldest
)

type ConsumeFunc func(msg *sarama.ConsumerMessage) error

type ConsumerConfig struct {
	Addrs         []string
	ListenTopics  []string
	GroupId       string
	InitialOffset int64
	Concurrency   int
}

type Consumer interface {
	ConsumerMsgAndBlock(ctx context.Context, f ConsumeFunc, concurrency bool)
}

func NewConsumer(cfg *ConsumerConfig) (Consumer, error) {
	if cfg.GroupId == "" || len(cfg.ListenTopics) < 1 || len(cfg.Addrs) < 1 {
		return nil, fmt.Errorf("kafka config invalid, GroupId:%s ListenTopics:%s Addrs:%v\n", cfg.GroupId, cfg.ListenTopics, cfg.Addrs)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Version = sarama.V1_1_0_0

	if cfg.InitialOffset != 0 {
		saramaConfig.Consumer.Offsets.Initial = cfg.InitialOffset
	} else {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	client, err := sarama.NewClient(cfg.Addrs, saramaConfig)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	consumerGroup, err := sarama.NewConsumerGroupFromClient(cfg.GroupId, client)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &kafkaConsumer{
		ListenTopics:  cfg.ListenTopics,
		ConsumerGroup: consumerGroup,
	}, nil
}
