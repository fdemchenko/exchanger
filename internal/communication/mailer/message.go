package mailer

import (
	"time"
)

type MessageType string

const (
	ExchangeRateUpdated   MessageType = "ExchangeRateUpdated"
	SendEmailNotification MessageType = "SendEmailNotification"
)

type MessageHeader struct {
	Type      MessageType `json:"messageType"`
	Timestamp time.Time   `json:"timestamp"`
}

type Message[T any] struct {
	MessageHeader
	Payload T `json:"payload"`
}
