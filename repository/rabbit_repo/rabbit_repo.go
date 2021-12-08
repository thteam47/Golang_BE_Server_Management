package rabbitrepo

type EmailsPublisher interface {
	Publish(body []byte, contentType string) error
}
type EmailsConsumer interface {
	StartConsumer(workerPoolSize int, exchange, queueName, bindingKey, consumerTag string) error
}
