package sqs_tools

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"net/http"
	"time"
)

type Client interface {
	SendMsg(ctx context.Context, msg *sqs.SendMessageInput) error
	SendBytesMsg(ctx context.Context, msg []byte) error
	ReceiveMsgAndBlock(ctx context.Context, f HandleMsgFunc, opts ...ReceiveMsgOption)
	SqsClient() *sqs.Client
}

type Config struct {
	AccessKeyId     string
	SecretAccessKey string
	QueueUrl        string
	Region          string
}

func NewClient(cfg *Config) Client {
	httpClient := http.DefaultClient
	httpClient.Timeout = 10 * time.Second

	cs := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyId,
		cfg.SecretAccessKey,
		"",
	)

	client := sqs.NewFromConfig(
		aws.Config{
			Region:      cfg.Region,
			Credentials: cs,
			HTTPClient:  httpClient,
		})

	return &sqsClient{
		sqsClient: client,
		queueUrl:  cfg.QueueUrl,
	}
}
