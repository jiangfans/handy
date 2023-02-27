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

func (consumer *kafkaConsumer) Run(ctx context.Context, f ConsumeFunc, concurrency bool) {
	log.Infof("😂😂😂start receive msg ...")

	var programQuitNormal bool
	defer func() {
		if e := recover(); e != nil {
			if programQuitNormal {
				log.Info("😊program quit normal")
			} else {
				log.Errorf("😭program quit with panic: %v", e)
			}
		}

		_ = consumer.ConsumerGroup.Close()
	}()

	go func() {
		select {
		case <-ctx.Done():
			programQuitNormal = true
			panic("👋ctx done, program quit")
		}
	}()

	var handler sarama.ConsumerGroupHandler

	if !concurrency {
		handler = NewOneByOneConsumerHandler(f)
	} else {
		// todo 实现并发处理消息
		panic("🈚️concurrency consume not implement!")
	}

	for {
		err := consumer.ConsumerGroup.Consume(ctx, consumer.ListenTopics, handler)
		if err != nil {
			log.Error("😭program quit with error: ", err.Error())
			return
		}
	}
}
