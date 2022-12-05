package kafka_tools

import (
	"github.com/Shopify/sarama"
	"github.com/jiangfans/handy/utils"
	log "github.com/sirupsen/logrus"
	"time"
)

type ConsumerGroupHandler struct {
	consumeFunc ConsumeFunc
}

func NewConsumerGroupHandler(consumeFunc ConsumeFunc) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		consumeFunc: consumeFunc,
	}
}

func (*ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (*ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
		log.Printf("Message topic:%q partition:%d offset:%d timestamp:%v\n", msg.Topic, msg.Partition, msg.Offset, msg.Timestamp)
		err = handler.consumeFunc(msg)
		if err != nil {
			// 有错误直接返回，避免丢消息，这里有可能堵塞消费，先👀下
			return nil
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}
