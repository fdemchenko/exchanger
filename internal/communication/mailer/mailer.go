package mailer

import "github.com/fdemchenko/exchanger/internal/communication"

const QueueName = "emails"

const (
	ExchangeRateUpdated   communication.MessageType = "ExchangeRateUpdated"
	SendEmailNotification communication.MessageType = "SendEmailNotification"
)

type ExchangeRateUpdatedEvent struct {
	Rate float32 `json:"rate"`
}

type SendEmailNotificationCommand struct {
	Email string `json:"email"`
}
