package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	DefaultSMTPPort                 = 25
	DefaultRabbitMQPort             = 5672
	DefaultMailerConnectionPoolSize = 3
)

func main() {
	conn, err := amqp.Dial(ServiceConfig.RabbitMQConnString)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	queue, err := ch.QueueDeclare(
		"emails",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	deliveries, err := ch.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Str("name", queue.Name).Msg("Mialer service started")

	for delivery := range deliveries {
		log.Debug().Str("Payload", string(delivery.Body)).Send()
	}
}
