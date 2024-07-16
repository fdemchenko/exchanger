package customers

import "github.com/fdemchenko/exchanger/internal/communication"

const QueueName = "emails"

const (
	CreateCustomerRequestQueue  = "CreateCustomerRequests"
	CreateCustomerResponseQueue = "CreateCustomerResponces"
)

const (
	CreateCustomerRequest  communication.MessageType = "CreateCustomerRequest"
	CustomerCreated        communication.MessageType = "CustomerCreated"
	CustomerCreationFailed communication.MessageType = "CustomerCreationFailed"
)

type CreateCustomerRequestPayload struct {
	Email          string `json:"email"`
	SubscriptionID int    `json:"id"`
}

type CustomerCreatedPayload struct {
	ID int `json:"id"`
}

type CustomerCreationFailedPayload struct {
	Error          string `json:"error"`
	SubscriptionID int    `json:"id"`
}
