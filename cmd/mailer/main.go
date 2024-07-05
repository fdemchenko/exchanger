package main

import (
	"encoding/json"
	"strconv"

	"github.com/fdemchenko/exchanger/internal/communication/mailer"
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

	mailerService := NewMailerService(ServiceConfig.SMTP)
	mailerService.StartWorkers(ServiceConfig.SMTP.ConnectionPoolSize)

	for delivery := range deliveries {
		message := mailer.Message[json.RawMessage]{}
		err := json.Unmarshal(delivery.Body, &message)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}

		log.Debug().Str("message type", message.Type.String()).
			Time("timestamp", message.Timestamp).
			Msg("New message received")

		ParseMessageType(mailerService, message)
	}
}

func ParseMessageType(mailerService *MailerService, message mailer.Message[json.RawMessage]) {
	switch message.Type {
	case mailer.ExchangeRateUpdated:
		rate64, err := strconv.ParseFloat(string(message.Payload), 32)
		if err != nil {
			log.Error().Err(err).Msg("Cannot parse rate message payload")
		}
		err = mailerService.UpdateCurrencyRateTemplates(float32(rate64))
		if err != nil {
			log.Error().Err(err).Send()
		}
	case mailer.SendEmailNotification:
		email, err := strconv.Unquote(string(message.Payload))
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
		mailerService.SendEmail(email)
	default:
		log.Error().Msg("Unknown message type")
	}
}
