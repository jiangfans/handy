package monitor

const (
	kafkaConsumeTotal    = "kafka_consume_total"
	kafkaConsumeTimeCost = "kafka_consume_time_cost"

	sqsConsumeTotal    = "sqs_consume_total"
	sqsConsumeTimeCost = "sqs_consume_time_cost"

	requestTotal      = "request_total"
	requestTimeCost   = "request_time_cost"
	requestErrorTotal = "request_error_total"

	sqsMessages          = "sqs_approximatenumberofmessages"
	sqsMessageDelayed    = "sqs_approximatenumberofmessagesdelayed"
	sqsMessageNotVisible = "sqs_approximatenumberofmessagesnotvisible"
)
