package main

import (
	"io"
	"net/http"
)

func getRate(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Exchange rate")
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Subscibed")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /rate", getRate)
	mux.HandleFunc("POST /subscribe", subscribe)

	http.ListenAndServe(":8080", mux)
}
