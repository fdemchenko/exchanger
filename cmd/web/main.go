package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fdemchenko/exchanger/internal/services"
)

type config struct {
	addr string
}

type application struct {
	cfg               config
	rateService       *services.RateService
	errorLog, infoLog *log.Logger
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8080", "http listen address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Lshortfile)
	app := application{
		cfg:         cfg,
		rateService: services.NewRateService(time.Hour),
		errorLog:    errorLog,
		infoLog:     infoLog,
	}

	infoLog.Println("Starting web server at " + app.cfg.addr)
	log.Fatalln(http.ListenAndServe(app.cfg.addr, app.routes()))
}
