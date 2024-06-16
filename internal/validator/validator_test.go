package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidEmails(t *testing.T) {
	testCases := []struct {
		name  string
		email string
	}{
		{name: "Hyphen in address", email: "some-email@address.com"},
		{name: "Dot in address", email: "some.email@address.com"},
		{name: "Multi domain", email: "someemail@address.edu.ua"},
		{name: "Hyphen in the end", email: "someemail-@address.ua"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := New()
			v.Check(IsValidEmail(tc.email), "email", "invalid email address")

			assert.True(t, v.IsValid())
		})
	}
}

func TestValidator_InvalidEmails(t *testing.T) {
	testCases := []struct {
		name  string
		email string
	}{
		{name: "No domain", email: "some-email@"},
		{name: "Hash symbol in email", email: "abc.def@mail#archive.com"},
		{name: "Without at sign", email: "someemailaddress.edu.ua"},
	}

	message := "invalid email address"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := New()
			v.Check(IsValidEmail(tc.email), "email", message)

			assert.False(t, v.IsValid())
			assert.Equal(t, message, v.Errors["email"])
		})
	}
}
