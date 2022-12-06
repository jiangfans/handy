package kafka_tools

import (
	"github.com/Shopify/sarama"
	"github.com/jiangfans/handy/utils"
	log "github.com/sirupsen/logrus"
	"time"
)

/*
	一条条消息消费，适用于不能丢消息的场景
*/

type OneByOneConsumerHandler struct {
	consumeFunc ConsumeFunc
}

func NewOneByOneConsumerHandler(consumeFunc ConsumeFunc) *OneByOneConsumerHandler {
	return &OneByOneConsumerHandler{
		consumeFunc: consumeFunc,
	}
}

func (*OneByOneConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (*OneByOneConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *OneByOneConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var msg *sarama.ConsumerMessage
	var err error

	defer func() {
		logFields := map[string]interface{}{
			"topic":     msg.Topic,
			"partition": msg.Partition,
			"offset":    msg.Offset,
			"time":      msg.Timestamp.In(utils.CST).Format(time.RFC3339),
			"key":       msg.Key,
			"value":     string(msg.Value),
		}

		if re := recover(); re != nil {
			if msg != nil {
				log.WithFields(logFields).Errorf("consume kafka msg paniced: %v", re)
			} else {
				log.Errorf("consume kafka msg paniced: %v", re)
			}
		}

		if err != nil {
			if msg != nil {
				log.WithFields(logFields).Errorf("consume kafka msg error: " + err.Error())
			} else {
				log.Errorf("consume kafka msg error: " + err.Error())
			}
		}
	}()

	for msg := range claim.Messages() {
		log.WithFields(log.Fields{
			"topic":     msg.Topic,
			"partition": msg.Partition,
			"offset":    msg.Offset,
			"timestamp": msg.Timestamp.In(utils.CST).Format(time.RFC3339),
		}).Debug("received msg")

		err = handler.consumeFunc(msg)
		if err != nil {
			// 有错误直接返回，避免丢消息，这里有可能堵塞消费，先👀下
			return nil
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}
