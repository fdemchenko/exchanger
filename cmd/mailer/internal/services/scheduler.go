package services

import (
	"time"

	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/mailer"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	"github.com/rs/zerolog/log"
)

type EmailScheduler struct {
	producer *rabbitmq.GenericProducer
	stopChan chan bool
}

func NewEmailScheduler(
	producer *rabbitmq.GenericProducer,
) *EmailScheduler {
	return &EmailScheduler{
		stopChan: make(chan bool),
		producer: producer,
	}
}

func (es *EmailScheduler) StartBackhroundTask(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				es.sendTriggerMessage()
			case <-es.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (es *EmailScheduler) sendTriggerMessage() {
	msg := communication.Message[struct{}]{
		MessageHeader: communication.MessageHeader{Type: mailer.StartEmailSending, Timestamp: time.Now()},
	}
	err := es.producer.SendMessage(msg, mailer.TriggerEmailsSendingQueue)
	if err != nil {
		log.Error().Err(err).Send()
	}
}

func (es *EmailScheduler) StopBackgroundTask() {
	es.stopChan <- true
}
