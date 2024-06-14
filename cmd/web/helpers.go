package main

import (
	"net/http"
	"regexp"

	"github.com/rs/zerolog/log"
)

//nolint:lll
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (app *application) serverError(w http.ResponseWriter, err error) {
	log.Error().Err(err).Send()
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func isCorrectEmail(email string) bool {
	return EmailRX.MatchString(email)
}
