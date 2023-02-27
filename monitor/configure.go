package monitor

import (
	"errors"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.shoplazza.site/xiabing/goat.git/prom"
)

var KafkaProm, RequestProm, RequestErrorProm *prom.PromVec

type Config struct {
	Namespace      string
	KafkaEnabled   bool
	RequestEnabled bool
}

func Configure(cfg *Config) error {
	if cfg == nil {
		return errors.New("config can't be nil")
	}

	if cfg.Namespace == "" {
		return errors.New("namespace can't be empty")
	}

	if cfg.KafkaEnabled {
		KafkaProm = prom.NewPromVec(cfg.Namespace).
			Counter(kafkaConsumeTotal, "Kafka consume total", []string{"topic", "result"}).
			Histogram(kafkaConsumeTimeCost, "Kafka consume time cost", []string{"topic"}, prometheus.ExponentialBuckets(0.02, 2, 11))
	}

	if cfg.RequestEnabled {
		RequestProm = prom.NewPromVec(cfg.Namespace).
			Counter(requestTotal, "Request total", []string{"url", "method", "status_code"}).
			Histogram(requestTimeCost, "Request time cost", []string{"url", "method"}, prometheus.ExponentialBuckets(0.02, 2, 11))

		RequestErrorProm = prom.NewPromVec(cfg.Namespace).
			Counter(requestErrorTotal, "Request error total", []string{"url", "method"})
	}

	return nil
}

func ReportKafkaConsumeTotal(topic, result string) {
	if KafkaProm != nil {
		KafkaProm.Inc(topic, result)
	}
}

func ReportKafkaConsumeTimeCost(startTime time.Time, topic string) {
	if KafkaProm != nil {
		KafkaProm.HandleTime(startTime, topic)
	}
}

func ReportRequestTotal(reqUrl, method string, statusCode int) {
	if RequestProm != nil {
		RequestProm.Inc(reqUrl, method, strconv.Itoa(statusCode))
	}
}

func ReportRequestErrorTotal(reqUrl, method string) {
	if RequestErrorProm != nil {
		RequestErrorProm.Inc(reqUrl, method)
	}
}

func ReportRequestTimeCost(startTime time.Time, reqUrl, method string) {
	if RequestProm != nil {
		RequestProm.HandleTime(startTime, reqUrl, method)
	}
}
