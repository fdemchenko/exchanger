package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func OpenWithQueueName(connectionString string, queueName string) (*amqp.Channel, error) {
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return ch, nil
}
