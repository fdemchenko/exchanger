package main

import "net/http"

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.errorLog.Println(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
