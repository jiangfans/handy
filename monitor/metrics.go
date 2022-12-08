package monitor

const (
	kafkaConsumeTotal    = "built_in_kafka_consume_total"
	kafkaConsumeTimeCost = "built_in_kafka_consume_time_cost"

	requestTotal      = "built_in_request_total"
	requestTimeCost   = "built_in_request_time_cost"
	requestErrorTotal = "built_in_request_error_total"

	sqsMessages          = "built_in_sqs_num_of_messages"
	sqsMessageDelayed    = "built_in_sqs_num_of_messages_delayed"
	sqsMessageNotVisible = "built_in_sqs_num_of_messages_not_visible"
)
