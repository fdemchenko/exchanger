package main

import (
	"github.com/fdemchenko/exchanger/cmd/mailer/internal/config"
	"github.com/fdemchenko/exchanger/cmd/mailer/internal/messaging"
	"github.com/fdemchenko/exchanger/cmd/mailer/internal/services"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/rs/zerolog/log"
)

const (
	DefaultMailerConnectionPoolSize = 3
)

func main() {
	cfg := config.LoadConfig()
	rateEmailsChannel, err := rabbitmq.OpenWithQueueName(cfg.RabbitMQConnString, mailer.RateEmailsQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	emailsTriggersChannel, err := rabbitmq.OpenWithQueueName(cfg.RabbitMQConnString, mailer.TriggerEmailsSendingQueue)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	mailerService := services.NewMailerService(cfg.SMTP)
	mailerService.StartWorkers(cfg.SMTP.ConnectionPoolSize)

	producer := rabbitmq.NewGenericProducer(emailsTriggersChannel)
	scheduler := services.NewEmailScheduler(producer)
	scheduler.StartBackhroundTask(cfg.SchedulerInterval)

	forever := make(chan bool)
	consumer := messaging.NewRateEmailsConsumer(rateEmailsChannel, mailerService)
	err = consumer.StartListening()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Msg("Mialer service started")
	<-forever
}
