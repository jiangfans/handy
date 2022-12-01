package sqs_tools

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type sqsClient struct {
	sqsClient *sqs.Client
	queueUrl  string
}

type ReceiveMsgOpts struct {
	MaxNumberOfMessages  int32           // 单词轮训获取到到最大消息数量
	VisibilityTimeout    int32           // 消息能被下一次查询查到的间隔时间
	WaitTimeSeconds      int32           // 轮训消息间隔时间
	HandleMsgConcurrency int             // 消费消息的并发goroutine数目
	RetryIntervals       []time.Duration // 数组下标为重试次数，值为重试间隔时间
}

const (
	DefaultMaxNumberOfMessages  = 10 // 10s
	DefaultWaitTimeSeconds      = 5  // 5s
	DefaultVisibilityTimeout    = 60 // 60s
	DefaultHandleMsgConcurrency = 5
)

type HandleMsgFunc func(ctx context.Context, msg *types.Message) error

type (
	funcReceiveMsgOption struct {
		f func(opts *ReceiveMsgOpts)
	}

	ReceiveMsgOption interface {
		apply(opts *ReceiveMsgOpts)
	}
)

func (fdo *funcReceiveMsgOption) apply(do *ReceiveMsgOpts) {
	fdo.f(do)
}

func newReceiveMsgOption(f func(opts *ReceiveMsgOpts)) *funcReceiveMsgOption {
	return &funcReceiveMsgOption{
		f: f,
	}
}

func MaxNumberOfMessages(number int32) ReceiveMsgOption {
	return newReceiveMsgOption(func(opts *ReceiveMsgOpts) {
		opts.MaxNumberOfMessages = number
	})
}

func VisibilityTimeout(seconds int32) ReceiveMsgOption {
	return newReceiveMsgOption(func(opts *ReceiveMsgOpts) {
		opts.VisibilityTimeout = seconds
	})
}

func WaitTimeSeconds(seconds int32) ReceiveMsgOption {
	return newReceiveMsgOption(func(opts *ReceiveMsgOpts) {
		opts.WaitTimeSeconds = seconds
	})
}

func RetryIntervals(intervals []time.Duration) ReceiveMsgOption {
	return newReceiveMsgOption(func(opts *ReceiveMsgOpts) {
		opts.RetryIntervals = intervals
	})
}

func HandleMsgConcurrency(concurrency int) ReceiveMsgOption {
	return newReceiveMsgOption(func(opts *ReceiveMsgOpts) {
		opts.HandleMsgConcurrency = concurrency
	})
}

func (sc *sqsClient) SendBytesMsg(ctx context.Context, msg []byte) error {
	sMInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(msg)),
		QueueUrl:    aws.String(sc.queueUrl),
	}

	_, err := sc.sqsClient.SendMessage(ctx, sMInput)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (sc *sqsClient) SendMsg(ctx context.Context, msg *sqs.SendMessageInput) error {
	_, err := sc.sqsClient.SendMessage(ctx, msg)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (sc *sqsClient) ReceiveMsgAndBlock(ctx context.Context, f HandleMsgFunc, opts ...ReceiveMsgOption) {
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
	}()

	go func() {
		select {
		case <-ctx.Done():
			programQuitNormal = true
			panic("ctx done, program quit!")
		}
	}()

	rMOpts := &ReceiveMsgOpts{}
	for _, opt := range opts {
		opt.apply(rMOpts)
	}

	gMInput := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(sc.queueUrl),
		AttributeNames:        []types.QueueAttributeName{types.QueueAttributeNameAll},
		MessageAttributeNames: []string{"All"},
	}

	gMInput.MaxNumberOfMessages = DefaultMaxNumberOfMessages
	if rMOpts.MaxNumberOfMessages != 0 {
		gMInput.MaxNumberOfMessages = rMOpts.MaxNumberOfMessages
	}

	gMInput.VisibilityTimeout = DefaultVisibilityTimeout
	if rMOpts.VisibilityTimeout != 0 {
		gMInput.VisibilityTimeout = rMOpts.VisibilityTimeout
	}

	gMInput.WaitTimeSeconds = DefaultWaitTimeSeconds
	if rMOpts.WaitTimeSeconds != 0 {
		gMInput.WaitTimeSeconds = rMOpts.WaitTimeSeconds
	}

	concurrency := DefaultHandleMsgConcurrency
	if rMOpts.HandleMsgConcurrency != 0 {
		concurrency = rMOpts.HandleMsgConcurrency
	}

	concurrencyChan := make(chan struct{}, concurrency)

	for {
		rMOutput, err := sc.sqsClient.ReceiveMessage(ctx, gMInput)
		if err != nil {
			log.Error(err.Error())
			time.Sleep(2 * time.Second)
		}

		for _, message := range rMOutput.Messages {
			concurrencyChan <- struct{}{}

			go func(msg types.Message) {
				defer func() {
					if e := recover(); e != nil {
						log.Error("handle sqs msg panic: ", e)
					}

					<-concurrencyChan
				}()

				sc.handleMsg(ctx, &msg, f, opts...)
			}(message)
		}
	}
}

func (sc *sqsClient) handleMsg(ctx context.Context, msg *types.Message, f HandleMsgFunc, opts ...ReceiveMsgOption) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("😭consume sqs msg panic: ", e)
		}
	}()

	rMOpts := &ReceiveMsgOpts{}
	for _, opt := range opts {
		opt.apply(rMOpts)
	}

	if err := f(ctx, msg); err != nil {
		log.Error(err.Error())

		// 如果需要重试，更改VisibilityTimeout
		// 获取已经接收到的消息次数
		if len(rMOpts.RetryIntervals) != 0 {
			if value, ok := msg.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)]; ok {
				receiveCount, err := strconv.Atoi(value)
				if err != nil {
					log.Error(err.Error())
					return
				}

				var retryTimes int
				if receiveCount > 1 {
					retryTimes = receiveCount - 1
				}

				if retryTimes < len(rMOpts.RetryIntervals)-1 {
					sc.changeMessageVisibility(ctx, msg, int32(rMOpts.RetryIntervals[retryTimes]/time.Second))
					return
				} else {
					log.Infof("retry over allow times")
				}
			}

			log.Error("msg attribute <ApproximateReceiveCount> not found")
		}
	}

	// 删除消息
	sc.deleteMessage(ctx, msg)
}

func (sc *sqsClient) changeMessageVisibility(ctx context.Context, msg *types.Message, visibilityTimeout int32) {
	cMVInput := sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(sc.queueUrl),
		ReceiptHandle:     msg.ReceiptHandle,
		VisibilityTimeout: visibilityTimeout,
	}

	_, err := sc.sqsClient.ChangeMessageVisibility(ctx, &cMVInput)
	if err != nil {
		log.Error(err.Error())
	}
	log.Debugf("message %s change visibility timeount to %d", aws.ToString(msg.MessageId), visibilityTimeout)
	return
}

func (sc *sqsClient) deleteMessage(ctx context.Context, msg *types.Message) {
	dMInput := sqs.DeleteMessageInput{
		QueueUrl:      aws.String(sc.queueUrl),
		ReceiptHandle: msg.ReceiptHandle,
	}

	_, err := sc.sqsClient.DeleteMessage(ctx, &dMInput)
	if err != nil {
		log.Error(err.Error())
	}
	log.Debugf("message %s deleted", aws.ToString(msg.MessageId))
	return
}

func (sc *sqsClient) SqsClient() *sqs.Client {
	return sc.sqsClient
}
