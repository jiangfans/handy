package monitor

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	log "github.com/sirupsen/logrus"
)

// Default to checking queues every 30 seconds
const defaultMonitorInterval = 30

func MonitorSQSQueue(sqsClient *sqs.Client, queueUrl string) {
	queueComponents := strings.Split(queueUrl, "/")
	queueName := queueComponents[len(queueComponents)-1]

	params := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueUrl),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
			types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
			types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
		},
	}

	for {
		resp, err := sqsClient.GetQueueAttributes(context.Background(), params)
		if err != nil {
			log.Error("GetQueueAttributes error: " + err.Error())
			continue
		}

		for attrib, prop := range resp.Attributes {
			nMessages, _ := strconv.ParseFloat(prop, 64)
			switch attrib {
			case string(types.QueueAttributeNameApproximateNumberOfMessages):
				SqsPromMessages.Set(nMessages, queueName)
			case string(types.QueueAttributeNameApproximateNumberOfMessagesDelayed):
				SqsPromMessagesDelayed.Set(nMessages, queueName)
			case string(types.QueueAttributeNameApproximateNumberOfMessagesNotVisible):
				SqsPromMessagesNotVisible.Set(nMessages, queueName)
			default:
				log.Errorf("unknown attribute %v", attrib)
			}
		}

		time.Sleep(defaultMonitorInterval * time.Second)
	}
}
