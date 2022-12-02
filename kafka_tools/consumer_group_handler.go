package kafka_tools

import (
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

type ConsumerGroupHandler struct {
	consumeFunc ConsumeFunc
}

func NewConsumerGroupHandler(consumeFunc ConsumeFunc) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{consumeFunc: consumeFunc}
}

func (*ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (*ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	defer func() {
		if msg != nil {
			CapturePanic(config.Cfg.SentryDSN, NewKafkaMessage(msg))
		} else {
			CapturePanic(config.Cfg.SentryDSN)
		}
	}()

	var msg *sarama.ConsumerMessage
	for msg = range claim.Messages() {
		log.Printf("Message topic:%q partition:%d offset:%d timestamp:%v\n", msg.Topic, msg.Partition, msg.Offset, msg.Timestamp)
		err := h.consumeFunc(msg)
		if err != nil {
			log.WithError(err).Warn("consumer err")
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
