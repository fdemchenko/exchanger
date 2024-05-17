package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fdemchenko/exchanger/models"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /rate", app.getRate)
	mux.HandleFunc("POST /subscribe", app.subscribe)

	return mux
}

func (app *application) getRate(w http.ResponseWriter, r *http.Request) {
	rate, err := app.rateService.GetRate()
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "%f", rate.Rates.UAH)
}

func (app *application) subscribe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	newEmail := r.PostForm.Get("email")
	if !isCorrectEmail(newEmail) {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	err = app.emailModel.Insert(newEmail)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			app.clientError(w, http.StatusConflict)
			return
		}
		app.serverError(w, err)
	}
}
