package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fdemchenko/exchanger/internal/models"
	"github.com/fdemchenko/exchanger/internal/services"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type config struct {
	addr string
	db   struct {
		dsn            string
		maxConnections int
	}
}

type application struct {
	cfg               config
	rateService       *services.RateService
	errorLog, infoLog *log.Logger
	emailModel        *models.EmailModel
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8080", "http listen address")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EXCHANGER_DSN"), "Data source name")
	flag.IntVar(&cfg.db.maxConnections, "db-max-conn", 25, "Database max connection")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Lshortfile)

	db, err := openDB(cfg)
	if err != nil {
		errorLog.Fatalln(err.Error())
	}
	infoLog.Println("Coonected to DB successfuly")

	app := application{
		cfg:         cfg,
		rateService: services.NewRateService(time.Hour),
		errorLog:    errorLog,
		infoLog:     infoLog,
		emailModel:  &models.EmailModel{DB: db},
	}

	infoLog.Println("Starting web server at " + app.cfg.addr)
	log.Fatalln(http.ListenAndServe(app.cfg.addr, app.routes()))
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxConnections)
	return db, nil
}
