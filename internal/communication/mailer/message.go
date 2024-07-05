package mailer

import (
	"errors"
	"strconv"
	"time"
)

type MessageType uint

const (
	ExchangeRateUpdated MessageType = iota
	SendEmailNotification
)

var MessageToString = map[MessageType]string{
	ExchangeRateUpdated:   "ExchangeRateUpdated",
	SendEmailNotification: "SendEmailNotification",
}

var MessageFromString = map[string]MessageType{
	"ExchangeRateUpdated":   ExchangeRateUpdated,
	"SendEmailNotification": SendEmailNotification,
}

func (mt MessageType) String() string {
	return MessageToString[mt]
}

func (mt MessageType) MarshalJSON() ([]byte, error) {
	quotedMessageType := strconv.Quote(mt.String())
	return []byte(quotedMessageType), nil
}

func (mt *MessageType) UnmarshalJSON(data []byte) error {
	messageTypeString, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	messageType, exists := MessageFromString[messageTypeString]
	if !exists {
		return errors.New("unknown message type")
	}
	*mt = messageType
	return nil
}

type MessageHeader struct {
	Type      MessageType `json:"messageType"`
	Timestamp time.Time   `json:"timestamp"`
}

type Message[T any] struct {
	MessageHeader
	Payload T `json:"payload"`
}
