package main

import (
	"fmt"
	"io"
	"net/http"
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
	io.WriteString(w, "Subscibed")
}
